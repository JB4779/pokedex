package pokecache

import (
	"bytes"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache stores response bytes for a fixed lifetime. Use NewCache to create one.
type Cache struct {
	mu       sync.Mutex
	entries  map[string]cacheEntry
	interval time.Duration
	done     chan struct{}
	once     sync.Once
}

// NewCache starts a background worker that removes expired entries.
// The interval must be positive. Call Close when the cache is no longer needed.
func NewCache(interval time.Duration) *Cache {
	if interval <= 0 {
		panic("cache interval must be positive")
	}
	c := &Cache{entries: make(map[string]cacheEntry), interval: interval, done: make(chan struct{})}
	go c.reapLoop()
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{createdAt: time.Now(), val: bytes.Clone(val)}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Since(entry.createdAt) >= c.interval {
		delete(c.entries, key)
		return nil, false
	}
	return bytes.Clone(entry.val), true
}

// Close stops background cleanup and is safe to call more than once.
func (c *Cache) Close() { c.once.Do(func() { close(c.done) }) }

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			c.mu.Lock()
			for key, entry := range c.entries {
				if now.Sub(entry.createdAt) >= c.interval {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		case <-c.done:
			return
		}
	}
}
