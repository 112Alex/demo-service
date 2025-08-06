package config

import (
	"os"
)

// Config содержит все настройки приложения.
type Config struct {
	DBUser       string
	DBPassword   string
	DBName       string
	DBHost       string
	DBPort       string
	KafkaBrokers []string
	KafkaTopic   string
	HTTPPort     string
}

// NewConfig загружает конфигурацию из переменных окружения.
func NewConfig() *Config {
	return &Config{
		DBUser:       getEnv("POSTGRES_USER", "test_user"),
		DBPassword:   getEnv("POSTGRES_PASSWORD", "test_password"),
		DBName:       getEnv("POSTGRES_DB", "orders_db"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKER", "localhost:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC", "orders"),
		HTTPPort:     getEnv("HTTP_PORT", "8081"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
