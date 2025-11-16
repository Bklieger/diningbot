package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached menu result
type CacheEntry struct {
	Items     []string
	Timestamp time.Time
}

// MenuCache provides thread-safe caching for menu results
type MenuCache struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry
	ttl   time.Duration
}

// NewMenuCache creates a new menu cache with the specified TTL
func NewMenuCache(ttl time.Duration) *MenuCache {
	return &MenuCache{
		items: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
}

// key generates a cache key from location, date, and mealType
func (c *MenuCache) key(location, date, mealType string) string {
	return location + "|" + date + "|" + mealType
}

// Get retrieves a cached menu if it exists and hasn't expired
func (c *MenuCache) Get(location, date, mealType string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.key(location, date, mealType)
	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}

	// Return a copy to prevent external modification
	result := make([]string, len(entry.Items))
	copy(result, entry.Items)
	return result, true
}

// Set stores a menu result in the cache
func (c *MenuCache) Set(location, date, mealType string, items []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.key(location, date, mealType)
	// Create a copy to prevent external modification
	copyItems := make([]string, len(items))
	copy(copyItems, items)

	c.items[key] = &CacheEntry{
		Items:     copyItems,
		Timestamp: time.Now(),
	}
}

// Clear removes all entries from the cache
func (c *MenuCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheEntry)
}

// CleanExpired removes expired entries from the cache
func (c *MenuCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.items {
		if now.Sub(entry.Timestamp) > c.ttl {
			delete(c.items, key)
		}
	}
}
