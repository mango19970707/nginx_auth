package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/coocood/freecache"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gateway/models"
)

var userStore *models.UserStore
var thirdPartyUrl string

const cacheSize = 100 * 1024 * 1024

var cache = freecache.NewCache(cacheSize)

// middleware to check if user is logged in
func authRequired(c *gin.Context) {
	userCookie, err := c.Cookie("user")
	if userCookie == "" || err != nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}
	token, err := cache.Get([]byte(userCookie))
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	} else {
		if tokenCookie, err := c.Cookie("sw-auth-token"); err != nil || string(token) != tokenCookie {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
	}
	cache.Set([]byte(userCookie), token, 900)
	c.Next()
}

// middleware to check user permission
func requirePermission(level int) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := c.Cookie("user")
		if username == "" || err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if user, _ := userStore.GetUser(username); user.Permission < level {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"message": "权限不足",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	// Initialize user store
	userStore = models.NewUserStore()

	// Get third-party URL from environment variable, or use default
	thirdPartyUrl = os.Getenv("THIRD_PARTY_URL")
	if thirdPartyUrl == "" {
		thirdPartyUrl = "https://example.com/third-party" // Default fallback URL
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	// Public routes
	router.GET("/login", func(c *gin.Context) {
		userCookie, err := c.Cookie("user")
		if userCookie == "" || err != nil {
			c.HTML(http.StatusUnauthorized, "login.html", nil)
			return
		}
		if token, err := cache.Get([]byte(userCookie)); err != nil {
			c.HTML(http.StatusUnauthorized, "login.html", nil)
			return
		} else {
			if tokenCookie, err := c.Cookie("sw-auth-token"); err != nil || string(token) != tokenCookie {
				c.HTML(http.StatusUnauthorized, "login.html", nil)
				return
			}
		}
		c.Redirect(http.StatusFound, "/dashboard")
	})

	router.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		user, exists := userStore.GetUser(username)
		if !exists || user.Password != password {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"error": "用户名或密码无效",
			})
			return
		}

		// Set cookies for authentication
		c.SetCookie("user", username, 900, "/", "", false, true)
		authToken := GenerateRandomString()
		c.SetCookie("sw-auth-token", authToken, 900, "/", "", false, true)
		cache.Set([]byte(username), []byte(authToken), 900)
		c.Redirect(http.StatusFound, "/dashboard")
	})

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(authRequired)
	{
		authorized.GET("/dashboard", func(c *gin.Context) {
			username, err := c.Cookie("user")
			if err != nil {
				c.HTML(http.StatusUnauthorized, "login.html", nil)
				return
			}
			user, exists := userStore.GetUser(username)
			if !exists {
				c.HTML(http.StatusUnauthorized, "login.html", nil)
				return
			}

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"username":      username,
				"permission":    user.Permission,
				"thirdPartyUrl": thirdPartyUrl,
			})
		})

		authorized.GET("/logout", func(c *gin.Context) {
			username, _ := c.Cookie("user")
			c.SetCookie("user", "", -1, "/", "", false, true)
			c.SetCookie("sw-auth-token", "", -1, "/", "", false, true)
			cache.Del([]byte(username))
			c.Redirect(http.StatusFound, "/login")
		})

		// User management routes
		userGroup := authorized.Group("/manage")
		userGroup.Use(requirePermission(2))
		{
			userGroup.GET("/list", func(c *gin.Context) {
				users := userStore.GetAllUsers()
				c.HTML(http.StatusOK, "user_list.html", gin.H{
					"Users": users,
				})
			})

			userGroup.GET("/add", func(c *gin.Context) {
				c.HTML(http.StatusOK, "user_form.html", gin.H{
					"User": &models.User{},
				})
			})

			userGroup.POST("/add", func(c *gin.Context) {
				username := c.PostForm("username")
				password := c.PostForm("password")
				permission, _ := strconv.Atoi(c.PostForm("permission"))

				user := &models.User{
					Username:   username,
					Password:   password, // Already hashed by client-side
					Permission: permission,
				}

				if userStore.AddUser(user) {
					c.Redirect(http.StatusFound, "/manage/list")
				} else {
					c.HTML(http.StatusOK, "user_list.html", gin.H{
						"User":  user,
						"Error": "Username already exists",
					})
				}
			})

			userGroup.GET("/edit/:username", func(c *gin.Context) {
				username := c.Param("username")
				if user, exists := userStore.GetUser(username); exists {
					c.HTML(http.StatusOK, "user_form.html", gin.H{
						"User": user,
					})
				} else {
					c.Redirect(http.StatusFound, "/manage/list")
				}
			})

			userGroup.POST("/edit/:username", func(c *gin.Context) {
				username := c.Param("username")
				password := c.PostForm("password")
				permission, _ := strconv.Atoi(c.PostForm("permission"))

				user := &models.User{
					Username:   username,
					Password:   password, // Already hashed by client-side
					Permission: permission,
				}

				if userStore.UpdateUser(user) {
					c.Redirect(http.StatusFound, "/manage/list")
				} else {
					c.HTML(http.StatusOK, "user_list.html", gin.H{
						"User":  user,
						"Error": "Failed to update user",
					})
				}
			})

			userGroup.POST("/delete/:username", func(c *gin.Context) {
				username := c.Param("username")
				userStore.DeleteUser(username)
				c.Redirect(http.StatusFound, "/manage/list")
			})
		}
	}

	router.Run(":28080")
}

func GenerateRandomString() string {
	// 生成48字节的随机数据（48*8 = 384位）
	bytes := make([]byte, 48)
	_, err := rand.Read(bytes)
	if err != nil {
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}
	// 使用URL安全的Base64编码，生成64字符的字符串
	return base64.URLEncoding.EncodeToString(bytes)
}
