package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/112Alex/demo-service.git/internal/cache"
	"github.com/112Alex/demo-service.git/internal/config"
	"github.com/112Alex/demo-service.git/internal/model"

	"github.com/segmentio/kafka-go"
)

type OrderSaver interface {
	SaveOrder(ctx context.Context, order *model.Order) error
}

// Consumer represents a Kafka consumer with retry and DLQ support.
// It consumes messages, attempts to persist them, stores them in cache and commits offsets.
type Consumer struct {
	reader *kafka.Reader
	writer *kafka.Writer
	db     OrderSaver
	cache  *cache.Cache

	maxRetries     int
	retryBackoff   time.Duration
}

// NewConsumer creates a consumer and DLQ producer based on config.
func NewConsumer(cfg *config.Config, dbSaver OrderSaver, cache *cache.Cache) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: "order-consumer-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.KafkaDeadTopic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Consumer{
		reader: reader,
		writer: writer,
		db:     dbSaver,
		cache:  cache,
		maxRetries:   cfg.KafkaMaxRetries,
		retryBackoff: cfg.KafkaRetryBackoff,
	}
}

// StartConsumption launches message consumption loop until context is cancelled.
func (c *Consumer) StartConsumption(ctx context.Context) {
	log.Println("Запуск потребителя Kafka...")

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("Ошибка FetchMessage: %v", err)
			continue
		}

		if err := c.handleMessage(ctx, m); err != nil {
			log.Printf("Не удалось обработать сообщение offset %d: %v", m.Offset, err)
			// don't commit offset to allow reprocessing; optionally send to DLQ already here
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("Ошибка CommitMessages: %v", err)
		}
	}

	_ = c.reader.Close()
	_ = c.writer.Close()
}

// handleMessage processes message with retry and DLQ.
func (c *Consumer) handleMessage(ctx context.Context, m kafka.Message) error {
	var order model.Order
	if err := json.Unmarshal(m.Value, &order); err != nil {
		log.Printf("Ошибка JSON, отправляем в DLQ: %v", err)
		return c.produceToDLQ(ctx, m)
	}
	if order.OrderUID == "" {
		log.Println("Пустой order_uid, отправляем в DLQ")
		return c.produceToDLQ(ctx, m)
	}

	// retry loop for DB save
	var err error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err = c.db.SaveOrder(ctx, &order)
		if err == nil {
			break
		}
		log.Printf("Ошибка сохранения заказа %s, попытка %d/%d: %v", order.OrderUID, attempt+1, c.maxRetries+1, err)
		time.Sleep(c.retryBackoff)
	}

	if err != nil {
		log.Printf("Не удалось сохранить заказ %s после %d попыток, отправка в DLQ", order.OrderUID, c.maxRetries+1)
		return c.produceToDLQ(ctx, m)
	}

	c.cache.Set(order.OrderUID, &order)
	return nil
}

// produceToDLQ sends the original message to dead-letter topic.
func (c *Consumer) produceToDLQ(ctx context.Context, m kafka.Message) error {
	return c.writer.WriteMessages(ctx, kafka.Message{Key: m.Key, Value: m.Value, Time: time.Now()})
}
