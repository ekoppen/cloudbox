package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
	BaseURL     string
	MasterKey   string // Master key for encrypting sensitive data
	
	// CORS settings
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	
	// Rate limiting
	RateLimitRequests int
	RateLimitWindow   string
	
	// File upload
	MaxFileSize   int64
	UploadPath    string
	AllowedTypes  []string
	
	// Backup settings
	BackupDir     string
}

// Load reads configuration from environment variables and config files
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("JWT_SECRET", "your-super-secret-jwt-key")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("BASE_URL", "http://localhost:8080")
	viper.SetDefault("MASTER_KEY", "") // Must be set in production
	
	// CORS defaults
	viper.SetDefault("ALLOWED_ORIGINS", []string{"*"})
	viper.SetDefault("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("ALLOWED_HEADERS", []string{"*"})
	
	// Rate limiting defaults
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", "1h")
	
	// File upload defaults
	viper.SetDefault("MAX_FILE_SIZE", 10<<20) // 10MB
	viper.SetDefault("UPLOAD_PATH", "./uploads")
	viper.SetDefault("ALLOWED_TYPES", []string{"image/jpeg", "image/png", "image/gif", "text/html", "text/css", "application/javascript"})
	
	// Backup defaults
	viper.SetDefault("BACKUP_DIR", "/var/lib/cloudbox/backups")

	// Bind environment variables
	viper.AutomaticEnv()

	maxFileSize, _ := strconv.ParseInt(getEnvOrDefault("MAX_FILE_SIZE", "10485760"), 10, 64)

	config := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		DatabaseURL: getEnvOrDefault("DATABASE_URL", ""),
		JWTSecret:   getEnvOrDefault("JWT_SECRET", "your-super-secret-jwt-key"),
		RedisURL:    getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		BaseURL:     getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		MasterKey:   getEnvOrDefault("MASTER_KEY", ""),
		
		AllowedOrigins: viper.GetStringSlice("ALLOWED_ORIGINS"),
		AllowedMethods: viper.GetStringSlice("ALLOWED_METHODS"),
		AllowedHeaders: viper.GetStringSlice("ALLOWED_HEADERS"),
		
		RateLimitRequests: viper.GetInt("RATE_LIMIT_REQUESTS"),
		RateLimitWindow:   viper.GetString("RATE_LIMIT_WINDOW"),
		
		MaxFileSize:  maxFileSize,
		UploadPath:   getEnvOrDefault("UPLOAD_PATH", "./uploads"),
		AllowedTypes: viper.GetStringSlice("ALLOWED_TYPES"),
		
		BackupDir:    getEnvOrDefault("BACKUP_DIR", "/var/lib/cloudbox/backups"),
	}

	return config, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}