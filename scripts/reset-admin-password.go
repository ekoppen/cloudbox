package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User represents the user model (simplified for this script)
type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Email        string `json:"email" gorm:"uniqueIndex;not null"`
	Name         string `json:"name" gorm:"not null"`
	PasswordHash string `json:"-" gorm:"column:password_hash;not null"`
	Role         string `json:"role" gorm:"default:'admin'"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
}

func main() {
	fmt.Println("🔑 CloudBox Admin Password Reset Tool")
	fmt.Println("=====================================")
	fmt.Println()

	// Get database connection from environment or prompt
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_CONNECTION_STRING")
	}
	if dbURL == "" {
		dbURL = "postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox?sslmode=disable"
		fmt.Printf("Using default database URL: %s\n", dbURL)
		fmt.Print("Press Enter to continue or Ctrl+C to abort...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	// Connect to database
	fmt.Println("📡 Connecting to database...")
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	fmt.Println("✅ Connected to database successfully")
	fmt.Println()

	// Get user input
	email := getEmailInput()
	password := getPasswordInput()
	
	// Confirm action
	if !confirmAction(email) {
		fmt.Println("❌ Operation cancelled")
		return
	}

	// Hash the password
	fmt.Println("🔐 Hashing password...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	// Find or create user
	var user User
	result := db.Where("email = ?", email).First(&user)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new admin user
		fmt.Printf("👤 User %s not found. Creating new admin user...\n", email)
		
		name := getNameInput(email)
		user = User{
			Email:        email,
			Name:         name,
			PasswordHash: string(hashedPassword),
			Role:         "admin",
			IsActive:     true,
		}
		
		if err := db.Create(&user).Error; err != nil {
			log.Fatalf("❌ Failed to create user: %v", err)
		}
		
		fmt.Printf("✅ Admin user '%s' created successfully\n", email)
	} else if result.Error != nil {
		log.Fatalf("❌ Database error: %v", result.Error)
	} else {
		// Update existing user
		fmt.Printf("👤 Found existing user: %s (%s)\n", user.Name, user.Email)
		
		// Update password and ensure admin role
		updates := map[string]interface{}{
			"password_hash": string(hashedPassword),
			"role":          "admin",
			"is_active":     true,
		}
		
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			log.Fatalf("❌ Failed to update user: %v", err)
		}
		
		fmt.Printf("✅ Password reset successfully for user '%s'\n", email)
		fmt.Printf("✅ User role set to 'admin'\n")
		fmt.Printf("✅ User account activated\n")
	}

	fmt.Println()
	fmt.Println("🎉 Operation completed successfully!")
	fmt.Printf("📧 Email: %s\n", user.Email)
	fmt.Printf("👤 Name: %s\n", user.Name)
	fmt.Printf("🛡️  Role: %s\n", user.Role)
	fmt.Printf("✅ Active: %t\n", user.IsActive)
	fmt.Printf("🆔 User ID: %d\n", user.ID)
	fmt.Println()
	fmt.Println("You can now login to CloudBox with these credentials.")
}

func getEmailInput() string {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("📧 Enter admin email address: ")
		email, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Error reading input: %v\n", err)
			continue
		}
		
		email = strings.TrimSpace(email)
		if email == "" {
			fmt.Println("❌ Email cannot be empty")
			continue
		}
		
		if !strings.Contains(email, "@") {
			fmt.Println("❌ Please enter a valid email address")
			continue
		}
		
		return email
	}
}

func getPasswordInput() string {
	for {
		fmt.Print("🔑 Enter new password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Printf("\n❌ Error reading password: %v\n", err)
			continue
		}
		fmt.Println() // New line after password input
		
		if len(password) < 6 {
			fmt.Println("❌ Password must be at least 6 characters long")
			continue
		}
		
		fmt.Print("🔑 Confirm new password: ")
		confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Printf("\n❌ Error reading password confirmation: %v\n", err)
			continue
		}
		fmt.Println() // New line after password input
		
		if string(password) != string(confirmPassword) {
			fmt.Println("❌ Passwords do not match")
			continue
		}
		
		return string(password)
	}
}

func getNameInput(email string) string {
	reader := bufio.NewReader(os.Stdin)
	
	// Extract name from email as default
	defaultName := strings.Split(email, "@")[0]
	defaultName = strings.Title(strings.ReplaceAll(defaultName, ".", " "))
	
	fmt.Printf("👤 Enter full name (default: %s): ", defaultName)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Error reading input, using default: %s\n", defaultName)
		return defaultName
	}
	
	name = strings.TrimSpace(name)
	if name == "" {
		return defaultName
	}
	
	return name
}

func confirmAction(email string) bool {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("⚠️  This will reset the password for '%s' and set role to 'admin'\n", email)
	fmt.Print("Are you sure you want to continue? (yes/no): ")
	
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Error reading input: %v\n", err)
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y"
}