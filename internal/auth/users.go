package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials is returned when username or password is incorrect
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrUserNotFound is returned when the user doesn't exist
	ErrUserNotFound = errors.New("user not found")
)

// User represents a user in the system
type User struct {
	ID           string
	Username     string
	PasswordHash string
}

// UserStore manages user data
type UserStore struct {
	users map[string]*User
}

// NewUserStore creates a new user store with predefined users
func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*User),
	}
}

// InitializeDefaultUsers adds default users to the store
// In production, this would load from a database or environment variables
func (s *UserStore) InitializeDefaultUsers() error {
	// Default users for development/testing
	// Passwords: "password123" for all users
	defaultUsers := []struct {
		id       string
		username string
		password string
	}{
		{"1", "sshetty", "Utd@Pogba6"},
		{"2", "ajain", "acdc@mumbai1"},
		{"3", "nsoundararaj", "ishva@coimbatore1"},
	}

	for _, u := range defaultUsers {
		hashedPassword, err := HashPassword(u.password)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", u.username, err)
		}

		s.users[u.username] = &User{
			ID:           u.id,
			Username:     u.username,
			PasswordHash: hashedPassword,
		}
	}

	return nil
}

// AddUser adds a new user to the store
func (s *UserStore) AddUser(id, username, password string) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	s.users[username] = &User{
		ID:           id,
		Username:     username,
		PasswordHash: hashedPassword,
	}

	return nil
}

// ValidateCredentials validates username and password
func (s *UserStore) ValidateCredentials(username, password string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, ErrInvalidCredentials
	}

	// Compare password with hash
	if err := ComparePassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUser retrieves a user by username
func (s *UserStore) GetUser(username string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	// Use bcrypt cost of 10 (default)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// ComparePassword compares a password with a hash
func ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password comparison failed: %w", err)
	}
	return nil
}

// Global user store instance
var globalUserStore *UserStore

// InitUserStore initializes the global user store
func InitUserStore() error {
	globalUserStore = NewUserStore()
	return globalUserStore.InitializeDefaultUsers()
}

// GetUserStore returns the global user store instance
func GetUserStore() *UserStore {
	if globalUserStore == nil {
		// Initialize if not already done
		_ = InitUserStore()
	}
	return globalUserStore
}

