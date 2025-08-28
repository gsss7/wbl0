package cache

import (
	"sync"
	"time"
)

type entry[T any] struct {
	v   T     //value
	exp int64 //expiration time
}

type Cache[T any] struct {
	mu   sync.RWMutex
	data map[string]entry[T]
	ttl  time.Duration
	max  int
	stop chan struct{}
}

func New[T any](ttl time.Duration, maxItems int) *Cache[T] {
	c := &Cache[T]{data: make(map[string]entry[T]), ttl: ttl, max: maxItems, stop: make(chan struct{})}
	go c.janitor()
	return c
}

func (c *Cache[T]) janitor() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixNano()
			c.mu.Lock()
			for k, e := range c.data {
				if e.exp > 0 && e.exp <= now {
					delete(c.data, k)
				}
			}
			if c.max > 0 && len(c.data) > c.max {
				overflow := len(c.data) - c.max
				i := 0
				for k := range c.data {
					delete(c.data, k)
					i++
					if i >= overflow {
						break
					}
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache[T]) Stop() { close(c.stop) }

func (c *Cache[T]) Set(key string, v T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp := time.Now().Add(c.ttl).UnixNano()
	c.data[key] = entry[T]{v: v, exp: exp}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	var zero T
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		return zero, false
	}
	if e.exp > 0 && time.Now().UnixNano() > e.exp {
		c.Invalidate(key)
		return zero, false
	}
	return e.v, true
}

func (c *Cache[T]) Invalidate(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *Cache[T]) Len() int { c.mu.RLock(); defer c.mu.RUnlock(); return len(c.data) }
