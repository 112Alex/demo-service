package cache

import (
    "container/list"
    "sync"
    "time"

    "github.com/112Alex/demo-service.git/internal/model"
)

// entry wraps cached order with metadata required for eviction.
// It is kept unexported to prevent direct mutations from outside the package.
// All access must happen through Cache methods to guarantee thread-safety.
type entry struct {
    key       string
    value     *model.Order
    timestamp time.Time     // creation time for TTL eviction
    element   *list.Element // pointer to underlying list node for O(1) moves
}

// Cache implements a fixed-capacity, thread-safe LRU cache with optional TTL eviction.
type Cache struct {
    mu       sync.RWMutex
    capacity int
    ttl      time.Duration

    items map[string]*entry // fast key lookup
    lru   *list.List        // doubly-linked list of *entry; most-recent at front
}

// NewCache returns a cache with the given capacity and TTL. A zero TTL disables time-based eviction.
func NewCache(capacity int, ttl time.Duration) *Cache {
    if capacity <= 0 {
        panic("capacity must be positive")
    }
    return &Cache{
        capacity: capacity,
        ttl:      ttl,
        items:    make(map[string]*entry, capacity),
        lru:      list.New(),
    }
}

// Set adds or updates an order in the cache.
// If the cache exceeds its capacity, the least-recently-used item is evicted.
func (c *Cache) Set(orderUID string, order *model.Order) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Update existing item if present
    if e, ok := c.items[orderUID]; ok {
        e.value = order
        e.timestamp = time.Now()
        c.lru.MoveToFront(e.element)
        return
    }

    // Insert new item
    e := &entry{key: orderUID, value: order, timestamp: time.Now()}
    e.element = c.lru.PushFront(e)
    c.items[orderUID] = e

    // Evict if over capacity
    if len(c.items) > c.capacity {
        c.evictOldest()
    }
}

// Get returns an order and true if found and not expired.
func (c *Cache) Get(orderUID string) (*model.Order, bool) {
    c.mu.RLock()
    e, ok := c.items[orderUID]
    c.mu.RUnlock()
    if !ok {
        return nil, false
    }

    // Check TTL without holding write lock first
    if c.ttl > 0 && time.Since(e.timestamp) > c.ttl {
        // expire item
        c.mu.Lock()
        c.removeElement(e)
        c.mu.Unlock()
        return nil, false
    }

    // Update recency
    c.mu.Lock()
    c.lru.MoveToFront(e.element)
    c.mu.Unlock()

    return e.value, true
}

// GetAll returns a shallow copy of all cached orders. Expired items are skipped.
func (c *Cache) GetAll() map[string]*model.Order {
    now := time.Now()
    result := make(map[string]*model.Order)

    c.mu.RLock()
    for k, e := range c.items {
        if c.ttl > 0 && now.Sub(e.timestamp) > c.ttl {
            continue
        }
        result[k] = e.value
    }
    c.mu.RUnlock()

    return result
}

// evictOldest removes the least-recently-used item.
func (c *Cache) evictOldest() {
    oldest := c.lru.Back()
    if oldest != nil {
        c.removeElement(oldest.Value.(*entry))
    }
}

// removeElement deletes an entry from both list and map. Caller must hold write lock.
func (c *Cache) removeElement(e *entry) {
    c.lru.Remove(e.element)
    delete(c.items, e.key)
}
