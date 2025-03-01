package rct

import (
	"sync"
	"time"
)

// An entry in a datagram cache (pair of datagram and timestamp)
type entry struct {
	dg *Datagram
	ts time.Time
}

// A datagram cache
type Cache struct {
	mu   sync.RWMutex
	data map[Identifier]entry
}

// Creates a new datagram cache
func NewCache() *Cache {
	return &Cache{data: make(map[Identifier]entry)}
}

// Returns cache entry for the given identifier
func (c *Cache) Get(id Identifier) (*Datagram, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry := c.data[id]
	return entry.dg, entry.ts
}

// Puts given datagram into the cache, for the identifier contained in the datagram
func (c *Cache) Put(dg *Datagram) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[dg.Id] = entry{dg, time.Now()}
}
