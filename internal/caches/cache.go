package caches

import (
	"WBTech_L0/internal/model"
	"WBTech_L0/internal/repository"
	"container/list"
	"errors"
	"fmt"
	"sync"
)

type Item struct {
	Key   string
	Value interface{}
}

type Cache struct {
	capacity int
	items    map[string]*list.Element
	queue    *list.List
	mu       sync.RWMutex
	db       *repository.Repository
}

func NewCache(db *repository.Repository, capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		queue:    list.New(),
		mu:       sync.RWMutex{},
		db:       db,
	}
}

func (c *Cache) Set(key string, value interface{}) error {
	c.mu.RLock()
	if element, ok := c.items[key]; ok == true {
		c.queue.MoveToFront(element)
		element.Value.(*Item).Value = value
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	if c.queue.Len() == c.capacity {
		c.purge()
	}

	item := &Item{
		Key:   key,
		Value: value,
	}

	element := c.queue.PushFront(item)

	c.mu.Lock()
	c.items[item.Key] = element
	c.mu.Unlock()

	return nil
}

func (c *Cache) Get(key string) interface{} {
	c.mu.RLock()
	element, ok := c.items[key]
	if ok == false {
		c.mu.RUnlock()
		return nil
	}

	c.queue.MoveToFront(element)
	c.mu.RUnlock()

	return element.Value.(*Item).Value
}

func (c *Cache) WarmCache() error {
	var orders []model.Order
	orders, err := c.db.GetOrdersForCache(c.capacity)
	if err != nil {
		return errors.New(fmt.Sprintf("Error in GetOrdersForCache() func: %s", err.Error()))
	}

	for _, order := range orders {
		if err = c.Set(order.OrderUID.String(), order); err != nil {
			return errors.New(fmt.Sprintf("Error while setting value for cache: %s", err.Error()))
		}
	}

	return nil
}

func (c *Cache) purge() {
	if element := c.queue.Back(); element != nil {
		item := c.queue.Remove(element).(*Item)
		delete(c.items, item.Key)
	}
}
