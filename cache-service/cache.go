package main

import (
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
)

type Cache struct {
	store *ristretto.Cache
}

func NewCache() *Cache {
	config := &ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30, // 1 GB RAM limit
		BufferItems: 64,
	}
	cache, err := ristretto.NewCache(config)
	if err != nil {
		log.Fatalf("failed to create cache: %v", err)
	}
	return &Cache{store: cache}
}

// put basically adds as and when required with a TTL
func (c *Cache) Put(key string, value string, ttl time.Duration) {
	c.store.SetWithTTL(key, value, 1, ttl)
	log.Printf("[CACHE] SET %s -> %s (ttl=%v)", key, value, ttl)
}

// get retrieves a key’s value if present.
func (c *Cache) Get(key string) (string, bool) {
	v, found := c.store.Get(key)
	if !found {
		log.Printf("[CACHE] MISS %s", key)
		return "", false
	}
	val := v.(string)
	log.Printf("[CACHE] HIT %s -> %s", key, val)
	return val, true
}

// invalidate deletes a key.
func (c *Cache) Invalidate(key string) {
	c.store.Del(key)
	log.Printf("[CACHE] DELETE %s", key)
}
