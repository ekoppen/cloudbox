package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	
	// GitHub integration
	GitHubToken   string
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
	viper.SetDefault("ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
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

	allowedMethods := getCORSMethods()
	fmt.Printf("DEBUG: ALLOWED_METHODS from env: %v\n", allowedMethods)
	
	config := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		DatabaseURL: getEnvOrDefault("DATABASE_URL", ""),
		JWTSecret:   getEnvOrDefault("JWT_SECRET", "your-super-secret-jwt-key"),
		RedisURL:    getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		BaseURL:     getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		MasterKey:   getEnvOrDefault("MASTER_KEY", ""),
		
		AllowedOrigins: getCORSOrigins(),
		AllowedMethods: allowedMethods,
		AllowedHeaders: getCORSHeaders(),
		
		RateLimitRequests: viper.GetInt("RATE_LIMIT_REQUESTS"),
		RateLimitWindow:   viper.GetString("RATE_LIMIT_WINDOW"),
		
		MaxFileSize:  maxFileSize,
		UploadPath:   getEnvOrDefault("UPLOAD_PATH", "./uploads"),
		AllowedTypes: viper.GetStringSlice("ALLOWED_TYPES"),
		
		BackupDir:    getEnvOrDefault("BACKUP_DIR", "/var/lib/cloudbox/backups"),
		GitHubToken:  getEnvOrDefault("GITHUB_TOKEN", ""),
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

// getStringSliceFromEnv gets comma-separated environment variable as string slice
func getStringSliceFromEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		fmt.Printf("DEBUG: Found env var %s = %s\n", key, value)
		result := strings.Split(value, ",")
		fmt.Printf("DEBUG: Split result: %v\n", result)
		return result
	}
	fmt.Printf("DEBUG: Using default for %s: %v\n", key, defaultValue)
	return defaultValue
}

// getCORSOrigins gets CORS origins from environment variables
// Tries CORS_ORIGINS first, then ALLOWED_ORIGINS, with sensible defaults
func getCORSOrigins() []string {
	// Try CORS_ORIGINS first (used by install script)
	if value := os.Getenv("CORS_ORIGINS"); value != "" {
		fmt.Printf("DEBUG: Found CORS_ORIGINS = %s\n", value)
		result := strings.Split(value, ",")
		// Trim whitespace from each origin
		for i, origin := range result {
			result[i] = strings.TrimSpace(origin)
		}
		fmt.Printf("DEBUG: CORS origins: %v\n", result)
		return result
	}
	
	// Fallback to ALLOWED_ORIGINS
	if value := os.Getenv("ALLOWED_ORIGINS"); value != "" {
		fmt.Printf("DEBUG: Found ALLOWED_ORIGINS = %s\n", value)
		result := strings.Split(value, ",")
		for i, origin := range result {
			result[i] = strings.TrimSpace(origin)
		}
		return result
	}
	
	// Default: allow all origins (development mode)
	fmt.Printf("DEBUG: Using default CORS origins: [*]\n")
	return []string{"*"}
}

// getCORSHeaders gets CORS headers from environment variables
// Tries CORS_HEADERS first, then ALLOWED_HEADERS, with sensible defaults
func getCORSHeaders() []string {
	// Try CORS_HEADERS first (preferred for CORS configuration)
	if value := os.Getenv("CORS_HEADERS"); value != "" {
		fmt.Printf("DEBUG: Found CORS_HEADERS = %s\n", value)
		result := strings.Split(value, ",")
		// Trim whitespace from each header
		for i, header := range result {
			result[i] = strings.TrimSpace(header)
		}
		fmt.Printf("DEBUG: CORS headers: %v\n", result)
		return result
	}
	
	// Fallback to ALLOWED_HEADERS (for backwards compatibility)
	if value := os.Getenv("ALLOWED_HEADERS"); value != "" {
		fmt.Printf("DEBUG: Found ALLOWED_HEADERS = %s\n", value)
		result := strings.Split(value, ",")
		for i, header := range result {
			result[i] = strings.TrimSpace(header)
		}
		return result
	}
	
	// Default: wildcard (will use comprehensive headers from middleware)
	fmt.Printf("DEBUG: Using default CORS headers: [*]\n")
	return []string{"*"}
}

// getCORSMethods gets CORS methods from environment variables
// Tries CORS_METHODS first, then ALLOWED_METHODS, with sensible defaults
func getCORSMethods() []string {
	// Try CORS_METHODS first (preferred for CORS configuration)
	if value := os.Getenv("CORS_METHODS"); value != "" {
		fmt.Printf("DEBUG: Found CORS_METHODS = %s\n", value)
		result := strings.Split(value, ",")
		// Trim whitespace from each method
		for i, method := range result {
			result[i] = strings.TrimSpace(method)
		}
		fmt.Printf("DEBUG: CORS methods: %v\n", result)
		return result
	}
	
	// Fallback to ALLOWED_METHODS (for backwards compatibility)
	if value := os.Getenv("ALLOWED_METHODS"); value != "" {
		fmt.Printf("DEBUG: Found ALLOWED_METHODS = %s\n", value)
		result := strings.Split(value, ",")
		for i, method := range result {
			result[i] = strings.TrimSpace(method)
		}
		return result
	}
	
	// Default: standard REST methods
	fmt.Printf("DEBUG: Using default CORS methods\n")
	return []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
}