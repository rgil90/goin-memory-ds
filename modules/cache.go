package modules

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64 // Unix timestamp in seconds for when this item expires (0 means no expiration)
}

type Cache struct {
	mutex          sync.RWMutex
	items       map[string]CacheItem
	stopCleanup chan bool
}

/*
This function creates a new cache with automatic cleanup.
*/
func NewCache() *Cache {
	cache := &Cache{
		items:       make(map[string]CacheItem),
		stopCleanup: make(chan bool),
	}
	
	// Start the background cleanup
	go cache.startCleanup()
	
	return cache
}

/*
This function adds a value to the cache with an optional TTL.

Parameters:
- key: The key of the item to set.
- value: The value to set in the cache.
- ttl: The time-to-live for the item. If not provided, the item will not expire.
*/
func (c *Cache) Set(key string, value interface{}, ttl ...time.Duration) {
	var expiration int64 = 0 // Default: no expiration
	
	if len(ttl) > 0 && ttl[0] > 0 {
		expiration = time.Now().Add(ttl[0]).Unix()
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

/*
This function retrieves a value from the cache, returning nil if expired or not found.
If the item is found and not expired, the item's TTL is extended to the new TTL.

Parameters:
- key: The key of the item to retrieve.
- value: The value to set in the cache.
- ttl: The time-to-live for the item. If not provided, the item will not expire.

Returns:
- The value of the item if found and not expired.
- False if the item is not found or expired.
*/
func (c *Cache) Get(key string) (interface{}, bool) {
	now := time.Now().Unix()

	c.mutex.RLock()
	item, found := c.items[key]
	c.mutex.RUnlock()
	
	if !found {
		return nil, false
	}
	
	// Check if the item has expired
	if item.Expiration > 0 && now >= item.Expiration {
		c.Delete(key)
		return nil, false
	}
	
	return item.Value, true
}

/*
This function deletes an item from the cache.
*/
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

/*
This function clears all items from the cache.
*/
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]CacheItem)
}

func (c *Cache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

/*
This function starts a ticker that removes expired items every second.
*/
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

/*
This function removes all expired items from the cache in two phases.
First, it identifies the expired keys.
Then, it deletes the expired keys from the cache.
This is done in two phases to avoid modifying the map during iteration.
*/
func (c *Cache) deleteExpired() {
	now := time.Now().Unix()
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	var expiredKeys []string
	
	// First, identify expired keys
	for key, item := range c.items {
		if item.Expiration > 0 && now >= item.Expiration {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(c.items, key)
	}
}

// Shut down the cache cleanup goroutine
func (c *Cache) Close() {
	close(c.stopCleanup)
}
