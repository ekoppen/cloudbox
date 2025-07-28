package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	apiKey := "a73cfd8f93c964e2eddd27e7827282d9c842278d7188f409fe13331775e4c060"
	storedHash := "$2b$12$damez8EzqArRGvfidw2BMOdwvE1ScbhVPTQqOVDdBhhl448fLxPZG"
	
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(apiKey))
	if err != nil {
		fmt.Printf("Hash verification failed: %v\n", err)
	} else {
		fmt.Println("Hash verification successful!")
	}
}