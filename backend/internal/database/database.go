package database

import (
	"fmt"
	"log"
	"time"

	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize creates a new database connection
func Initialize(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Migrate runs all database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Organization{},
		&models.OrganizationAdmin{},
		&models.Project{},
		&models.ProjectGitHubConfig{},
		&models.APIKey{},
		&models.CORSConfig{},
		&models.SSHKey{},
		&models.WebServer{},
		&models.GitHubRepository{},
		&models.Deployment{},
		&models.Backup{},
		&models.Collection{},
		&models.Document{},
		&models.Bucket{},
		&models.File{},
		&models.AppUser{},
		&models.AppSession{},
		&models.Channel{},
		&models.ChannelMember{},
		&models.Message{},
		&models.MessageReaction{},
		&models.MessageRead{},
		&models.Function{},
		&models.FunctionExecution{},
		&models.FunctionDomain{},
		&models.AuditLog{},
		&models.SystemSetting{},
		&utils.HostKeyEntry{}, // Add host key management
	)
}

// CreateDefaultSuperAdmin creates a default superadmin user if none exists
func CreateDefaultSuperAdmin(db *gorm.DB) error {
	// Check if any superadmin user already exists
	var count int64
	if err := db.Model(&models.User{}).Where("role = ?", "superadmin").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check for existing superadmin: %w", err)
	}

	// If superadmin already exists, skip creation
	if count > 0 {
		log.Printf("SuperAdmin user already exists, skipping creation")
		return nil
	}

	// Hash the default password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create default superadmin user
	defaultAdmin := models.User{
		Email:        "admin@cloudbox.local",
		Name:         "CloudBox Admin",
		PasswordHash: string(passwordHash),
		Role:         "superadmin",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&defaultAdmin).Error; err != nil {
		return fmt.Errorf("failed to create default superadmin: %w", err)
	}

	log.Printf("Default SuperAdmin user created: admin@cloudbox.local")
	return nil
}