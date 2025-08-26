package kafka

import (
    "context"
    "encoding/json"
    "errors"
    "testing"
    "time"

    "github.com/112Alex/demo-service.git/internal/cache"
    "github.com/112Alex/demo-service.git/internal/config"
    "github.com/112Alex/demo-service.git/internal/model"

    "github.com/segmentio/kafka-go"
)

type mockDB struct {
    saveErrCount int
}

func (m *mockDB) SaveOrder(ctx context.Context, o *model.Order) error {
    if m.saveErrCount > 0 {
        m.saveErrCount--
        return errors.New("temporary")
    }
    return nil
}

// remaining methods to satisfy interface (compile only)
func (m *mockDB) Close() error                        { return nil }

func TestConsumer_RetryLogic(t *testing.T) {
    cfg := &config.Config{KafkaMaxRetries: 2, KafkaRetryBackoff: 1 * time.Millisecond, CacheCapacity: 10, CacheTTL: 0}
    c := &Consumer{db: &mockDB{saveErrCount: 2}, cache: cache.NewCache(10, 0), maxRetries: cfg.KafkaMaxRetries, retryBackoff: cfg.KafkaRetryBackoff}
    order := model.Order{OrderUID: "1"}
    msgValue, _ := json.Marshal(order)
    err := c.handleMessage(context.Background(), kafka.Message{Value: msgValue})
    if err != nil {
        t.Errorf("expected success after retries, got %v", err)
    }
}