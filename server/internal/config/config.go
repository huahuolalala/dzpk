package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort string
	WSPort   string
	LogLevel string
	DBPath   string
}

func Load() *Config {
	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8081"),
		WSPort:   getEnv("WS_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		DBPath:   getEnv("DB_PATH", "./dz-poker.db"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
