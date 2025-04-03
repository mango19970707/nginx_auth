package models

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
)

// User represents a user in the system
type User struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`   // Stored as SHA-256 hash in Base64 format
	Permission int    `json:"permission"` // 1 or 2
}

// UserStore manages user data
type UserStore struct {
	Users map[string]*User
	file  string
}

// NewUserStore creates a new user store with initial admin user
func NewUserStore() *UserStore {
	store := &UserStore{
		Users: make(map[string]*User),
		file:  "data/users.json",
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(store.file), 0755); err != nil {
		panic(err)
	}

	// Try to load existing users from file
	if err := store.loadUsers(); err != nil {
		// If file doesn't exist or is invalid, create initial admin user
		store.Users["asguarder"] = &User{
			ID:         1,
			Username:   "asguarder",
			Password:   hashPassword("Sw@123456"), // This will be hashed in main.go before use
			Permission: 2,
		}
		// Save initial users to file
		if err := store.saveUsers(); err != nil {
			panic(err)
		}
	}

	return store
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// loadUsers loads users from the JSON file
func (s *UserStore) loadUsers() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.Users)
}

// saveUsers saves users to the JSON file
func (s *UserStore) saveUsers() error {
	data, err := json.MarshalIndent(s.Users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}

// GetUser retrieves a user by username
func (s *UserStore) GetUser(username string) (*User, bool) {
	user, exists := s.Users[username]
	return user, exists
}

// AddUser adds a new user
func (s *UserStore) AddUser(user *User) bool {
	if _, exists := s.Users[user.Username]; exists {
		return false
	}

	// Find highest ID to assign next ID
	var maxID uint = 0
	for _, u := range s.Users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}
	user.ID = maxID + 1

	s.Users[user.Username] = user
	if err := s.saveUsers(); err != nil {
		// If save fails, remove the user from memory
		delete(s.Users, user.Username)
		return false
	}
	return true
}

// UpdateUser updates an existing user
func (s *UserStore) UpdateUser(user *User) bool {
	if _, exists := s.Users[user.Username]; !exists {
		return false
	}

	// Preserve the original ID
	originalID := s.Users[user.Username].ID
	user.ID = originalID

	s.Users[user.Username] = user
	if err := s.saveUsers(); err != nil {
		return false
	}
	return true
}

// DeleteUser deletes a user by username
func (s *UserStore) DeleteUser(username string) bool {
	if _, exists := s.Users[username]; !exists {
		return false
	}

	delete(s.Users, username)
	if err := s.saveUsers(); err != nil {
		// If save fails, restore the user in memory
		s.Users[username] = s.Users[username]
		return false
	}
	return true
}

// GetAllUsers returns all users
func (s *UserStore) GetAllUsers() []*User {
	users := make([]*User, 0, len(s.Users))
	for _, user := range s.Users {
		users = append(users, user)
	}
	return users
}
