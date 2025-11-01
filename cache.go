package cache

import (
	"order-service/internal/domain"
	"sync"
)

type InMemoryCache struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		orders: make(map[string]*domain.Order),
	}
}

func (c *InMemoryCache) Set(orderUID string, order *domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[orderUID] = order
}

func (c *InMemoryCache) Get(orderUID string) (*domain.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.orders[orderUID]
	return order, exists
}

func (c *InMemoryCache) GetAll() map[string]*domain.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*domain.Order)
	for k, v := range c.orders {
		result[k] = v
	}
	return result
}

func (c *InMemoryCache) Restore(orders map[string]*domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders = orders
}
