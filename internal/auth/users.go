package auth

import (
	"errors"
	"fmt"
	"log"
	"os"

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
// Passwords are loaded from environment variables for security
func (s *UserStore) InitializeDefaultUsers() error {
	// Default users - passwords must be set via environment variables
	// Set USER_SSHETTY_PASSWORD, USER_AJAIN_PASSWORD, USER_NSOUNDARARAJ_PASSWORD
	defaultUsers := []struct {
		id       string
		username string
		envVar   string
	}{
		{"1", "sshetty", "USER_SSHETTY_PASSWORD"},
		{"2", "ajain", "USER_AJAIN_PASSWORD"},
		{"3", "nsoundararaj", "USER_NSOUNDARARAJ_PASSWORD"},
	}

	usersAdded := 0
	for _, u := range defaultUsers {
		password := os.Getenv(u.envVar)
		if password == "" {
			log.Printf("Warning: %s not set, skipping user %s", u.envVar, u.username)
			continue
		}

		hashedPassword, err := HashPassword(password)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", u.username, err)
		}

		s.users[u.username] = &User{
			ID:           u.id,
			Username:     u.username,
			PasswordHash: hashedPassword,
		}
		usersAdded++
	}

	if usersAdded == 0 {
		log.Printf("Warning: No default users were created. Set environment variables (USER_SSHETTY_PASSWORD, USER_AJAIN_PASSWORD, USER_NSOUNDARARAJ_PASSWORD) to create users.")
	} else {
		log.Printf("Initialized %d default user(s)", usersAdded)
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
