package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         int
	DBUser         string
	DBPass         string
	DBName         string
	PrivateKeyPath string
	AppPort        int
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("[WARN] .env dosyası okunamadı: %v", err)
	}

	cfg := &Config{}
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	cfg.DBPort = getEnvAsInt("DB_PORT", 5432)
	cfg.DBUser = getEnv("DB_USER", "postgres")
	cfg.DBPass = getEnv("DB_PASS", "postgres")
	cfg.DBName = getEnv("DB_NAME", "license_db")
	cfg.PrivateKeyPath = getEnv("PRIVATE_KEY_PATH", "./private_rsa.pem")
	cfg.AppPort = getEnvAsInt("APP_PORT", 4000)
	return cfg
}

func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	strVal := os.Getenv(key)
	if strVal == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}
