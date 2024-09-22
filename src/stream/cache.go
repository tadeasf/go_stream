package stream

import (
	"sync"
	"time"
)

type Cache struct {
	sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      []byte
	expiration int64
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
	}
}

func (c *Cache) Set(key string, value []byte, duration time.Duration) {
	c.Lock()
	defer c.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = cacheItem{
		value:      value,
		expiration: expiration,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}

	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}

func (c *Cache) Cleanup() {
	c.Lock()
	defer c.Unlock()

	for k, v := range c.items {
		if time.Now().UnixNano() > v.expiration {
			delete(c.items, k)
		}
	}
}
