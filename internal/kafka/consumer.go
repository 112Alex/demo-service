package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/112Alex/demo-service.git/internal/cache"
	"github.com/112Alex/demo-service.git/internal/db"
	"github.com/112Alex/demo-service.git/internal/model"

	"github.com/segmentio/kafka-go"
)

// Consumer представляет собой потребителя Kafka.
type Consumer struct {
	reader *kafka.Reader
	db     *db.DBClient
	cache  *cache.Cache
}

// NewConsumer создает и возвращает новый Kafka-потребитель.
func NewConsumer(brokers []string, topic string, db *db.DBClient, cache *cache.Cache) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "order-consumer-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader: reader,
		db:     db,
		cache:  cache,
	}
}

// StartConsumption начинает чтение сообщений из Kafka.
func (c *Consumer) StartConsumption(ctx context.Context) {
	log.Println("Запуск потребителя Kafka...")

	// Горутина для обработки сообщений
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Остановка потребителя Kafka...")
				c.reader.Close()
				return
			default:
				m, err := c.reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Ошибка при чтении сообщения: %v", err)
					continue
				}

				log.Printf("Получено сообщение из Kafka: %s", string(m.Value))

				// Десериализация сообщения
				var order model.Order
				if err := json.Unmarshal(m.Value, &order); err != nil {
					log.Printf("Ошибка при парсинге JSON: %v", err)
					continue // Пропускаем невалидное сообщение
				}

				// Валидация данных
				if order.OrderUID == "" {
					log.Println("Пропускаем сообщение с пустым order_uid")
					continue
				}

				// Сохранение в БД
				if err := c.db.SaveOrder(ctx, &order); err != nil {
					log.Printf("Ошибка при сохранении заказа %s в БД: %v", order.OrderUID, err)
					continue // Пропускаем, если не удалось сохранить
				}

				// Добавление в кэш
				c.cache.Set(order.OrderUID, &order)
				log.Printf("Заказ %s успешно обработан и сохранен", order.OrderUID)
			}
		}
	}()

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Получен сигнал завершения. Остановка...")
}
