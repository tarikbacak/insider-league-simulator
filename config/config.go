package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config structure holds the application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
}

// DatabaseConfig holds the database connection details
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

// ServerConfig holds the server configuration details
type ServerConfig struct {
	Port string `json:"port"`
}

// Global config variable for Singleton pattern
var AppConfig *Config

// Init loads the configuration file and initializes the global variable
func Init() {
	// Load .env file - if it fails, default values will be used
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, default values will be used")
	}

	// Create the configuration object
	AppConfig = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "insider_league"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}

	log.Println("Configuration successfully loaded")
}

// getEnv reads an environment variable, returns the default value if not found
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseURL constructs the PostgreSQL connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
	)
}

// GetConfig returns the global configuration object
func GetConfig() *Config {
	if AppConfig == nil {
		log.Fatal("Configuration not initialized. Call config.Init() function.")
	}
	return AppConfig
}
