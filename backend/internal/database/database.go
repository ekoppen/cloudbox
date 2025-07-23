package database

import (
	"fmt"

	"github.com/cloudbox/backend/internal/models"
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
		&models.Project{},
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
	)
}