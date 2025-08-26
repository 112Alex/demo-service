package cache

import (
	"testing"
	"time"

	"github.com/112Alex/demo-service.git/internal/model"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(10, 0)
	
	order := &model.Order{
		OrderUID: "test-123",
		TrackNumber: "TRACK123",
	}
	
	// Тест Set
	cache.Set(order.OrderUID, order)
	
	// Тест Get
	retrievedOrder, found := cache.Get(order.OrderUID)
	if !found {
		t.Error("Заказ не найден в кэше")
	}
	
	if retrievedOrder.OrderUID != order.OrderUID {
		t.Errorf("Ожидался OrderUID %s, получен %s", order.OrderUID, retrievedOrder.OrderUID)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := NewCache(10, 0)
	
	_, found := cache.Get("non-existent")
	if found {
		t.Error("Заказ не должен быть найден")
	}
}

func TestCache_GetAll(t *testing.T) {
	cache := NewCache(10, 0)
	
	order1 := &model.Order{OrderUID: "test-1", TrackNumber: "TRACK1"}
	order2 := &model.Order{OrderUID: "test-2", TrackNumber: "TRACK2"}
	
	cache.Set(order1.OrderUID, order1)
	cache.Set(order2.OrderUID, order2)
	
	allOrders := cache.GetAll()
	
	if len(allOrders) != 2 {
		t.Errorf("Ожидалось 2 заказа, получено %d", len(allOrders))
	}
	
	if _, exists := allOrders[order1.OrderUID]; !exists {
		t.Error("Первый заказ не найден в GetAll")
	}
	
	if _, exists := allOrders[order2.OrderUID]; !exists {
		t.Error("Второй заказ не найден в GetAll")
	}
}

func TestCache_EvictionLRU(t *testing.T) {
    c := NewCache(2, 0)
    o1 := &model.Order{OrderUID: "1"}
    o2 := &model.Order{OrderUID: "2"}
    o3 := &model.Order{OrderUID: "3"}

    c.Set(o1.OrderUID, o1)
    c.Set(o2.OrderUID, o2)
    _ = mustGet(t, c, o1.OrderUID) // access o1 to make it MRU

    c.Set(o3.OrderUID, o3) // should evict o2

    if _, ok := c.Get(o2.OrderUID); ok {
        t.Error("expected o2 to be evicted")
    }
    if _, ok := c.Get(o1.OrderUID); !ok {
        t.Error("expected o1 to stay")
    }
    if _, ok := c.Get(o3.OrderUID); !ok {
        t.Error("expected o3 to be present")
    }
}

func TestCache_TTLExpiration(t *testing.T) {
    c := NewCache(2, 10*time.Millisecond)
    o := &model.Order{OrderUID: "ttl"}
    c.Set(o.OrderUID, o)
    time.Sleep(20 * time.Millisecond)
    if _, ok := c.Get(o.OrderUID); ok {
        t.Error("expected item to expire by TTL")
    }
}

func mustGet(t *testing.T, c *Cache, uid string) *model.Order {
    t.Helper()
    v, ok := c.Get(uid)
    if !ok {
        t.Fatalf("uid %s not found", uid)
    }
    return v
}