package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	PrivateKeyPath string
	PublicKeyPath  string
	AESKeyPath     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Port:           getEnv("PORT", "3000"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "licensemanager"),
		PrivateKeyPath: getEnv("PRIVATE_KEY_PATH", "./keys/private_key.pem"),
		PublicKeyPath:  getEnv("PUBLIC_KEY_PATH", "./keys/public_key.pem"),
		AESKeyPath:     getEnv("AES_KEY_PATH", "./keys/aes.key"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
