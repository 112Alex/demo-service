package config

import (
	"fmt"
	"os"
	"strings"
	"time"
	"strconv"
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
	KafkaDeadTopic string
	HTTPPort     string
	// Cache settings
	CacheCapacity int
	CacheTTL      time.Duration
	// Kafka retry settings
	KafkaMaxRetries int
	KafkaRetryBackoff time.Duration
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
		KafkaDeadTopic: getEnv("KAFKA_DEAD_TOPIC", "orders-dlq"),
		HTTPPort:     getEnv("HTTP_PORT", "8081"),
		CacheCapacity: getEnvAsInt("CACHE_CAPACITY", 1000),
		CacheTTL:      getEnvAsDuration("CACHE_TTL", 10*time.Minute),
		KafkaMaxRetries: getEnvAsInt("KAFKA_MAX_RETRIES", 3),
		KafkaRetryBackoff: getEnvAsDuration("KAFKA_RETRY_BACKOFF", 500*time.Millisecond),
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
	if c.CacheCapacity <= 0 {
		return fmt.Errorf("CACHE_CAPACITY must be positive")
	}
	if c.CacheTTL < 0 {
		return fmt.Errorf("CACHE_TTL cannot be negative")
	}
	if c.KafkaDeadTopic == "" {
		return fmt.Errorf("KAFKA_DEAD_TOPIC не может быть пустым")
	}
	if c.KafkaMaxRetries < 0 {
		return fmt.Errorf("KAFKA_MAX_RETRIES cannot be negative")
	}
	if c.KafkaRetryBackoff < 0 {
		return fmt.Errorf("KAFKA_RETRY_BACKOFF cannot be negative")
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

func getEnvAsInt(key string, defaultVal int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
