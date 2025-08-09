package config

import (
	"fmt"
	"os"
	"strings"
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
	cfg := &Config{
		DBUser:       getEnv("POSTGRES_USER", "test_user"),
		DBPassword:   getEnv("POSTGRES_PASSWORD", "test_password"),
		DBName:       getEnv("POSTGRES_DB", "orders_db"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKER", "localhost:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "orders"),
		HTTPPort:     getEnv("HTTP_PORT", "8081"),
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("Ошибка валидации конфигурации: %v", err))
	}

	return cfg
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.DBUser == "" {
		return fmt.Errorf("POSTGRES_USER не может быть пустым")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("POSTGRES_PASSWORD не может быть пустым")
	}
	if c.DBName == "" {
		return fmt.Errorf("POSTGRES_DB не может быть пустым")
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST не может быть пустым")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DB_PORT не может быть пустым")
	}
	if len(c.KafkaBrokers) == 0 {
		return fmt.Errorf("KAFKA_BROKER не может быть пустым")
	}
	if c.KafkaTopic == "" {
		return fmt.Errorf("KAFKA_TOPIC не может быть пустым")
	}
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT не может быть пустым")
	}
	return nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
