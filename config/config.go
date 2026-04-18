package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string
	AWSAccessKey string
	AWSSecretKey string
	AWSRegion    string
	AWSBucket    string
}

func Load() *Config {
	// Load .env file in development (ignored if file not found)
	godotenv.Load()

	return &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "readme"),
		DBPassword:   getEnv("DB_PASSWORD", "readme123"),
		DBName:       getEnv("DB_NAME", "ms1_users"),
		JWTSecret:    getEnv("JWT_SECRET", "readme_jwt_secret_compartido_2026"),
		Port:         getEnv("PORT", "8001"),
		AWSAccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
		AWSBucket:    getEnv("AWS_S3_BUCKET", ""),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Lima",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
