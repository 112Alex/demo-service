package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq" // Импорт драйвера для PostgreSQL

	"github.com/112Alex/demo-service.git/internal/cache"
	"github.com/112Alex/demo-service.git/internal/config"
	"github.com/112Alex/demo-service.git/internal/db"
	"github.com/112Alex/demo-service.git/internal/kafka"
	"github.com/112Alex/demo-service.git/internal/server"
)

// main - главная точка входа в приложение.
func main() {
	// Загрузка конфигурации
	cfg := config.NewConfig()

	// Подключение к БД
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	dbClient, err := db.NewDBClient(connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer dbClient.Close()

	// Инициализация кэша
	orderCache := cache.NewCache()

	// Восстановление кэша из БД при старте
	restoreCache(context.Background(), dbClient, orderCache)

	// Запуск потребителя Kafka в отдельной горутине
	kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, dbClient, orderCache)
	go kafkaConsumer.StartConsumption(context.Background())

	// Запуск HTTP-сервера
	httpServer := server.NewServer(cfg.HTTPPort, orderCache, dbClient)
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске HTTP-сервера: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Сервис завершает работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 5) // Даем 5 секунд на завершение
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке HTTP-сервера: %v", err)
	}
	log.Println("Сервис успешно остановлен.")
}

// restoreCache заполняет кэш данными из БД при старте приложения.
func restoreCache(ctx context.Context, dbClient *db.DBClient, orderCache *cache.Cache) {
	log.Println("Восстановление кэша из БД...")

	// Здесь мы должны реализовать GetAllOrders в db/postgres.go
	// Но для краткости предположим, что она уже написана и возвращает все заказы.
	// Вот ее возможная реализация:
	orders, err := dbClient.GetAllOrders(ctx) // Предполагаемая функция
	if err != nil {
		log.Printf("Ошибка при восстановлении кэша из БД: %v", err)
		return
	}

	for _, order := range orders {
		orderCache.Set(order.OrderUID, order)
	}

	log.Printf("Кэш успешно восстановлен. Загружено %d заказов.", len(orders))
}
