package cache

import (
	"container/list"
	"runtime"
	"sync"
)

const MaxMemoryUsage = (14 * 1024 * 1024 * 1024) / 100 

type entry struct {
	key   string
	value string
}

type Cache struct {
	store map[string]*list.Element
	order *list.List
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		store: make(map[string]*list.Element),
		order: list.New(),
	}
}

func (c *Cache) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.store[key]; exists {
		c.order.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	elem := c.order.PushFront(&entry{key, value})
	c.store[key] = elem

	c.evictIfNeeded()
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, exists := c.store[key]; exists {
		c.order.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return "", false
}

func (c *Cache) evictIfNeeded() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for memStats.Alloc > MaxMemoryUsage {

		backElem := c.order.Back()
		if backElem == nil {
			return
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.store, backElem.Value.(*entry).key)
		c.order.Remove(backElem)

		runtime.ReadMemStats(&memStats)
	}
}

