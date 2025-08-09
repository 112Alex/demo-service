package cache

import (
	"testing"

	"github.com/112Alex/demo-service.git/internal/model"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache()
	
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
	cache := NewCache()
	
	_, found := cache.Get("non-existent")
	if found {
		t.Error("Заказ не должен быть найден")
	}
}

func TestCache_GetAll(t *testing.T) {
	cache := NewCache()
	
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