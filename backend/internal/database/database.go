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
	// First, run migrations for tables that don't have foreign key constraints
	if err := db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Organization{},
		&models.OrganizationAdmin{},
	); err != nil {
		return fmt.Errorf("failed to migrate basic tables: %w", err)
	}

	// Create default organization and handle project migration
	if err := MigrateProjectsWithOrganization(db); err != nil {
		return fmt.Errorf("failed to migrate projects with organization: %w", err)
	}

	// Continue with the rest of the migrations 
	// Note: We skip creating unique constraints for projects.slug
	// because we use a partial unique index (migration 010)
	if err := db.AutoMigrate(
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
		// Plugin management models
		&models.PluginRegistry{},
		&models.PluginInstallation{},
		&models.PluginState{},
		&models.ApprovedRepository{},
		&models.PluginDownload{},
		&models.PluginAuditLog{},
		&models.PluginSubmission{},
		&models.RepositoryApprovalRequest{},
		&models.PluginMarketplace{},
	); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	// Post-migration: Apply NOT NULL constraint after data migration
	if err := ApplyProjectOrganizationConstraints(db); err != nil {
		return fmt.Errorf("failed to apply organization constraints: %w", err)
	}

	// Post-migration: Remove any full unique constraint on slug that GORM might have created
	if err := EnsurePartialSlugConstraint(db); err != nil {
		return fmt.Errorf("failed to ensure partial slug constraint: %w", err)
	}

	return nil
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

// MigrateProjectsWithOrganization handles the migration of projects to require organizations
func MigrateProjectsWithOrganization(db *gorm.DB) error {
	// Check if projects table exists
	if !db.Migrator().HasTable(&models.Project{}) {
		log.Printf("Projects table doesn't exist yet, skipping migration")
		return nil
	}

	// First, create a default organization if no organizations exist
	var orgCount int64
	if err := db.Model(&models.Organization{}).Count(&orgCount).Error; err != nil {
		return fmt.Errorf("failed to count organizations: %w", err)
	}

	var defaultOrg models.Organization
	if orgCount == 0 {
		// Get the first user to be owner of default organization
		var firstUser models.User
		if err := db.First(&firstUser).Error; err != nil {
			log.Printf("No users exist yet, creating organization without user reference")
			// Create without user reference initially
			defaultOrg = models.Organization{
				Name:        "Default Organization",
				Description: "Default organization for existing projects",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
		} else {
			defaultOrg = models.Organization{
				Name:        "Default Organization", 
				Description: "Default organization for existing projects",
				UserID:      firstUser.ID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
		}

		if err := db.Create(&defaultOrg).Error; err != nil {
			return fmt.Errorf("failed to create default organization: %w", err)
		}
		log.Printf("Created default organization with ID: %d", defaultOrg.ID)
	} else {
		// Get the first organization
		if err := db.First(&defaultOrg).Error; err != nil {
			return fmt.Errorf("failed to get default organization: %w", err)
		}
	}

	// Check if organization_id column exists
	if !db.Migrator().HasColumn(&models.Project{}, "organization_id") {
		log.Printf("organization_id column doesn't exist yet, will be created in main migration")
		return nil
	}

	// Count projects without organization
	var projectCount int64
	if err := db.Model(&models.Project{}).Where("organization_id IS NULL OR organization_id = 0").Count(&projectCount).Error; err != nil {
		log.Printf("Could not count projects without organization: %v", err)
		return nil
	}

	// Update all projects without organization_id to use the default organization
	if projectCount > 0 {
		if err := db.Model(&models.Project{}).Where("organization_id IS NULL OR organization_id = 0").Update("organization_id", defaultOrg.ID).Error; err != nil {
			return fmt.Errorf("failed to update projects with default organization: %w", err)
		}
		log.Printf("Updated %d existing projects to use default organization", projectCount)
	}

	return nil
}

// ApplyProjectOrganizationConstraints applies NOT NULL constraint after data migration
func ApplyProjectOrganizationConstraints(db *gorm.DB) error {
	// Check if any projects still don't have organization_id
	var nullOrgCount int64
	if err := db.Model(&models.Project{}).Where("organization_id IS NULL OR organization_id = 0").Count(&nullOrgCount).Error; err != nil {
		log.Printf("Could not check for null organization_id: %v", err)
		return nil // Don't fail the migration
	}

	if nullOrgCount > 0 {
		log.Printf("Warning: %d projects still don't have organization_id set", nullOrgCount)
		return nil // Don't apply constraint yet
	}

	// Apply NOT NULL constraint using raw SQL
	if err := db.Exec("ALTER TABLE projects ALTER COLUMN organization_id SET NOT NULL").Error; err != nil {
		// This might fail if constraint already exists, which is fine
		log.Printf("Could not apply NOT NULL constraint to organization_id: %v", err)
	} else {
		log.Printf("Successfully applied NOT NULL constraint to projects.organization_id")
	}

	return nil
}

// EnsurePartialSlugConstraint ensures we have the correct partial unique constraint and removes full constraints
func EnsurePartialSlugConstraint(db *gorm.DB) error {
	// Drop any full unique constraint that GORM might have created
	constraints := []string{"idx_projects_slug", "projects_slug_key", "projects_slug_idx"}
	
	for _, constraint := range constraints {
		// Try to drop as constraint first
		if err := db.Exec("ALTER TABLE projects DROP CONSTRAINT IF EXISTS " + constraint).Error; err != nil {
			log.Printf("Could not drop constraint %s: %v", constraint, err)
		}
		
		// Try to drop as index
		if err := db.Exec("DROP INDEX IF EXISTS " + constraint).Error; err != nil {
			log.Printf("Could not drop index %s: %v", constraint, err)
		}
	}

	// Ensure our partial unique index exists
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'projects' AND indexname = 'idx_projects_slug_unique_active'").Scan(&count).Error; err != nil {
		log.Printf("Could not check for partial unique index: %v", err)
		return nil
	}

	if count == 0 {
		// Create the partial unique index if it doesn't exist
		if err := db.Exec("CREATE UNIQUE INDEX idx_projects_slug_unique_active ON projects(slug) WHERE deleted_at IS NULL").Error; err != nil {
			log.Printf("Could not create partial unique index: %v", err)
		} else {
			log.Printf("Created partial unique index idx_projects_slug_unique_active")
		}
	}

	return nil
}