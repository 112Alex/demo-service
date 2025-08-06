package cache

import (
	"log"
	"sync"

	"github.com/112Alex/demo-service.git/internal/model"
)

// Cache представляет собой потокобезопасный кэш для хранения заказов.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

// NewCache создает и возвращает новый экземпляр кэша.
func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
	}
}

// Set добавляет или обновляет заказ в кэше.
func (c *Cache) Set(orderUID string, order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[orderUID] = order
	log.Printf("Заказ %s добавлен в кэш", orderUID)
}

// Get извлекает заказ из кэша по orderUID.
// Возвращает nil, если заказ не найден.
func (c *Cache) Get(orderUID string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, found := c.orders[orderUID]
	return order, found
}

// GetAll возвращает все заказы из кэша.
func (c *Cache) GetAll() map[string]*model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Создаем копию, чтобы избежать изменений извне
	ordersCopy := make(map[string]*model.Order, len(c.orders))
	for k, v := range c.orders {
		ordersCopy[k] = v
	}
	return ordersCopy
}
