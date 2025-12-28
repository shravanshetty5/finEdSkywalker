package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	fmt.Println("=== Password Hash Generator for finEdSkywalker ===")
	fmt.Println()

	// Get username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading username: %v\n", err)
		os.Exit(1)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Fprintf(os.Stderr, "Username cannot be empty\n")
		os.Exit(1)
	}

	// Get password (hidden input)
	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
		os.Exit(1)
	}
	fmt.Println() // New line after password input

	password := string(passwordBytes)
	if password == "" {
		fmt.Fprintf(os.Stderr, "Password cannot be empty\n")
		os.Exit(1)
	}

	// Confirm password
	fmt.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError reading password confirmation: %v\n", err)
		os.Exit(1)
	}
	fmt.Println() // New line after password input

	confirmPassword := string(confirmBytes)
	if password != confirmPassword {
		fmt.Fprintf(os.Stderr, "Passwords do not match\n")
		os.Exit(1)
	}

	// Generate hash
	fmt.Println("\nGenerating password hash...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating password hash: %v\n", err)
		os.Exit(1)
	}

	// Display results
	fmt.Println("\n=== Generated Credentials ===")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password Hash: %s\n", string(hashedPassword))
	fmt.Println("\n=== Go Code Snippet ===")
	fmt.Printf("{\"%s\", \"%s\"},\n", username, string(hashedPassword))
	fmt.Println("\n=== Add to internal/auth/users.go ===")
	fmt.Println("Copy the Go code snippet above and add it to the defaultUsers array")
	fmt.Println()
}
