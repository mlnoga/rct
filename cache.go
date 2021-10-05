package rct

import (
	"time"
)

// An entry in a datagram cache (pair of datagram and timestamp)
type cacheEntry struct {
	dg *Datagram
	ts time.Time
}

// A datagram cache
type Cache struct {
	entries map[Identifier]cacheEntry
	timeout time.Duration
}

// Creates a new datagram cache
func NewCache(timeout time.Duration) (cache *Cache) {
	return &Cache{
		make(map[Identifier]cacheEntry),
		timeout,
	}
}

// Returns cache entry for the given identifier, if still valid under timeout
func (c *Cache) Get(i Identifier) (dg *Datagram, ok bool) {
	entry, ok := c.entries[i]
	if !ok || c.timeout < time.Since(entry.ts) {
		return &Datagram{}, false
	}
	return entry.dg, true
}

// Puts given datagram into the cache, for the identifier contained in the datagram
func (c *Cache) Put(dg *Datagram) {
	c.entries[dg.Id] = cacheEntry{dg, time.Now()}
}
