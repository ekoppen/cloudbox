package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin123"
	
	// Generate a fresh hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}
	
	fmt.Printf("Generated hash for '%s': %s\n", password, string(hash))
	
	// Test the fresh hash
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Printf("Fresh hash verification FAILED: %v\n", err)
	} else {
		fmt.Printf("Fresh hash verification SUCCESS\n")
	}
	
	// Test the hash we've been using
	oldHash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	err = bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(password))
	if err != nil {
		fmt.Printf("Old hash verification FAILED: %v\n", err)
	} else {
		fmt.Printf("Old hash verification SUCCESS\n")
	}
	
	// Test another known good hash
	knownGoodHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye/Vs7oHZGokfIXbJPCbJRx6M0JcKwxCu"
	err = bcrypt.CompareHashAndPassword([]byte(knownGoodHash), []byte(password))
	if err != nil {
		fmt.Printf("Known good hash verification FAILED: %v\n", err)
	} else {
		fmt.Printf("Known good hash verification SUCCESS\n")
	}
}