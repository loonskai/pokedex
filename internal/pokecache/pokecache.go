package pokecache

import (
	"sync"
	"time"
)

type cacheEntry[V any] struct {
	val       V
	createdAt time.Time
}

type Cache[V any] struct {
	entries map[string]cacheEntry[V]
	mu      sync.Mutex
}

func (cache *Cache[V]) Add(key string, val V) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry := cacheEntry[V]{
		val:       val,
		createdAt: time.Now(),
	}
	cache.entries[key] = entry
}

func (cache *Cache[V]) Get(key string) (V, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry, ok := cache.entries[key]
	if !ok {
		return *new(V), false
	}
	return entry.val, true
}

func NewCache[V any](interval time.Duration) *Cache[V] {
	cache := Cache[V]{
		entries: map[string]cacheEntry[V]{},
		mu:      sync.Mutex{},
	}
	go cache.reapLoop(interval)
	return &cache
}

func (cache *Cache[V]) reap(interval time.Duration) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for key, val := range cache.entries {
		if time.Since(val.createdAt) >= interval {
			delete(cache.entries, key)
		}
	}
}

func (cache *Cache[V]) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		cache.reap(interval)
	}
}
