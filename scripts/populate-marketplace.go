package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://cloudbox:cloudbox123@localhost:5432/cloudbox_db?sslmode=disable"
		log.Printf("Using default database URL (set DATABASE_URL environment variable to override)")
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Connected to database successfully")

	// Read the migration file
	migrationPath := filepath.Join("..", "backend", "migrations", "012_populate_marketplace_test_data.sql")
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute the migration
	log.Printf("Executing marketplace data population...")
	if err := db.Exec(string(migrationSQL)).Error; err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	log.Printf("✅ Marketplace data populated successfully!")

	// Verify the data was inserted
	var registryCount int64
	var marketplaceCount int64
	var approvedRepoCount int64

	db.Table("plugin_registry").Count(&registryCount)
	db.Table("plugin_marketplace").Count(&marketplaceCount)
	db.Table("approved_repositories").Count(&approvedRepoCount)

	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("- Plugin Registry entries: %d\n", registryCount)
	fmt.Printf("- Marketplace entries: %d\n", marketplaceCount)
	fmt.Printf("- Approved repositories: %d\n", approvedRepoCount)

	// Show the plugins that were added
	type PluginInfo struct {
		Name        string
		Version     string
		Author      string
		Category    string
		IsApproved  bool
		InstallCount int
	}

	var plugins []PluginInfo
	db.Table("plugin_registry pr").
		Select("pr.name, pr.version, pr.author, COALESCE(pm.category, 'Unknown') as category, pr.is_approved, pr.install_count").
		Joins("LEFT JOIN plugin_marketplace pm ON pr.name = pm.plugin_name").
		Where("pr.is_approved = ?", true).
		Find(&plugins)

	fmt.Printf("\n🔌 Available Plugins:\n")
	for _, plugin := range plugins {
		status := "✅"
		if !plugin.IsApproved {
			status = "⏳"
		}
		fmt.Printf("%s %s v%s by %s (%s) - %d installs\n", 
			status, plugin.Name, plugin.Version, plugin.Author, plugin.Category, plugin.InstallCount)
	}

	fmt.Printf("\n🌐 Test the marketplace at: http://localhost:8080/api/v1/marketplace/plugins\n")
}