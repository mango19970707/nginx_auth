package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coocood/freecache"

	"github.com/gin-gonic/gin"

	"gateway/models"
)

var userStore *models.UserStore

const cacheSize = 100 * 1024 * 1024

var cache = freecache.NewCache(cacheSize)

// middleware to check if user is logged in
func authRequired(c *gin.Context) {
	userCookie, err := c.Cookie("user")
	if userCookie == "" || err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", nil)
		c.Abort()
		return
	}
	token, err := cache.Get([]byte(userCookie))
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", nil)
		c.Abort()
		return
	} else {
		if tokenCookie, err := c.Cookie("sw-auth-token"); err != nil || string(token) != tokenCookie {
			c.HTML(http.StatusUnauthorized, "login.html", nil)
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
			c.HTML(http.StatusUnauthorized, "login.html", nil)
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

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/auth_static", "./static")

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

		user, exists := userStore.GetUser(userCookie)
		if !exists {
			c.HTML(http.StatusUnauthorized, "login.html", nil)
			return
		}

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"username":   userCookie,
			"permission": user.Permission,
		})
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
				"username":   username,
				"permission": user.Permission,
			})
		})

		authorized.GET("/logout", func(c *gin.Context) {
			username, _ := c.Cookie("user")
			c.SetCookie("user", "", -1, "/", "", false, true)
			c.SetCookie("sw-auth-token", "", -1, "/", "", false, true)
			cache.Del([]byte(username))
			c.HTML(http.StatusUnauthorized, "login.html", nil)
		})

		// ApiPost 路由
		authorized.GET("/apipost", func(c *gin.Context) {
			c.HTML(http.StatusOK, "apipost.html", nil)
		})

		authorized.POST("/apipost", func(c *gin.Context) {
			var request struct {
				URL    string            `json:"url"`
				Params map[string]string `json:"params"`
			}

			if err := c.BindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "无效的请求参数",
				})
				return
			}
			fmt.Println(request)

			// 创建HTTP客户端
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			// 准备数据
			data, err := json.Marshal(request.Params)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "无效的请求参数",
				})
				return
			}

			// 发送POST请求
			resp, err := client.Post(strings.TrimSpace(request.URL), "application/json", bytes.NewReader(data))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "请求失败",
					"message": err.Error(),
				})
				return
			}
			defer resp.Body.Close()

			// 读取响应内容
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "读取响应失败",
					"message": err.Error(),
				})
				return
			}

			// 尝试解析JSON响应
			var jsonResponse interface{}
			if err := json.Unmarshal(body, &jsonResponse); err != nil {
				// 如果不是JSON，返回原始响应
				c.JSON(http.StatusOK, gin.H{
					"status":   resp.StatusCode,
					"response": string(body),
				})
				return
			}

			// 返回JSON响应
			c.JSON(http.StatusOK, gin.H{
				"status":   resp.StatusCode,
				"response": jsonResponse,
			})
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

	listenPort := os.Getenv("GATEWAY_PORT")
	if _, err := strconv.Atoi(listenPort); err != nil {
		listenPort = "28080"
	}
	router.Run(":" + listenPort)
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
