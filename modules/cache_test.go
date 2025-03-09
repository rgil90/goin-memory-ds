package modules

import (
	"testing"
	"sync"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
This function tests the basic functionality of setting values and retrieving them from the cache,
including different value types and handling of non-existent keys.
*/
func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	// Test setting and retrieving a string
	cache.Set("string", "value")
	value, found := cache.Get("string")
	assert.True(t, found, "Expected to find key 'string'")
	assert.Equal(t, "value", value, "Expected value 'value'")
	
	// Test setting and retrieving an int
	cache.Set("int", 42)
	value, found = cache.Get("int")
	assert.True(t, found, "Expected to find key 'int'")
	assert.Equal(t, 42, value, "Expected value 42")
	
	// Test retrieving a non-existent key
	_, found = cache.Get("non-existent")
	assert.False(t, found, "Expected not to find key 'non-existent'")
}

/*
This function tests that items can be properly deleted from the cache,
and that deleting non-existent keys is handled gracefully.
*/
func TestCacheDelete(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	// Set a value
	cache.Set("key", "value")
	
	_, found := cache.Get("key")
	require.True(t, found, "Key was not set properly before deletion test")
	
	// Delete the key
	cache.Delete("key")
	
	_, found = cache.Get("key")
	assert.False(t, found, "Key was not deleted properly")
	
	// No panic should occur for non-existent key
	cache.Delete("non-existent")
}

/*
This function tests that the Clear method properly removes all items
from the cache, resetting it to an empty state.
*/
func TestCacheClear(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	// Add some data
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// Verify data exists
	if len(cache.Keys()) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(cache.Keys()))
	}
	
	// Clear the cache
	cache.Clear()
	
	// Verify it's empty
	if len(cache.Keys()) != 0 {
		t.Errorf("Expected 0 keys after clear, got %d", len(cache.Keys()))
	}
}

/*
This function tests that the Keys method correctly returns all keys
in the cache, including handling of empty caches.
*/
func TestCacheKeys(t *testing.T) {
	cache := NewCache()
	defer cache.Close()

	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected empty keys slice, got %v", keys)
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	for _, key := range expectedKeys {
		cache.Set(key, key+"_value")
	}

	keys = cache.Keys()

	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}
	
	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key %s not found in keys %v", expectedKey, keys)
		}
	}
}

/*
This function tests that the cache can handle concurrent
reads and writes from multiple goroutines without data corruption,
testing the thread safety of the implementation.
*/
func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache()
	defer cache.Close()

	const goroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Writers + Readers

	// Launch writer goroutines
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Set(key, j)
			}
		}(i)
	}
	
	// Launch reader goroutines
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key) // Don't need to check value, just testing concurrency
			}
		}(i)
	}
	
	wg.Wait()
	
	keys := cache.Keys()
	expectedItems := goroutines * operationsPerGoroutine
	if len(keys) != expectedItems {
		t.Errorf("Expected %d items in cache, got %d", expectedItems, len(keys))
	}
}

/*
This function tests edge cases such as nil values and empty string keys
to ensure the cache handles these cases correctly.
*/
func TestCacheEdgeCases(t *testing.T) {
	cache := NewCache()
	defer cache.Close()

	cache.Set("nil-key", nil)
	value, found := cache.Get("nil-key")
	if !found {
		t.Error("Expected to find key 'nil-key', but it was not found")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}

	cache.Set("", "empty-key")
	value, found = cache.Get("")
	if !found {
		t.Error("Expected to find empty string key, but it was not found")
	}
	if value != "empty-key" {
		t.Errorf("Expected 'empty-key' value, got %v", value)
	}
}

/*
This function tests that items with TTL are properly expired and
removed from the cache after their TTL has elapsed.
*/
func TestCacheTTL(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	cache.Set("short-lived", "value", 2*time.Second)
	
	value, found := cache.Get("short-lived")
	if !found {
		t.Error("Expected to find key right after setting it")
	}
	if value != "value" {
		t.Errorf("Expected 'value', got %v", value)
	}
	
	time.Sleep(3 * time.Second)
	
	_, found = cache.Get("short-lived")
	if found {
		t.Error("Expected item to be gone after TTL expired")
	}
}

/*
This function tests how the cache handles a mix of items with and
without TTL, ensuring that only the expired items are removed.
*/
func TestCacheMixedTTL(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	// Set an item with a TTL
	cache.Set("expires", "value1", 2*time.Second)
	
	// Set an item without a TTL
	cache.Set("permanent", "value2")
	
	// Wait for expiration of the first item
	time.Sleep(3 * time.Second)
	
	// Check that only the expiring item is gone
	_, found := cache.Get("expires")
	if found {
		t.Error("Expected 'expires' to be gone after TTL expired")
	}
	
	value, found := cache.Get("permanent")
	if !found {
		t.Error("Expected 'permanent' to still exist")
	}
	if value != "value2" {
		t.Errorf("Expected 'value2', got %v", value)
	}
}

/*
This function tests that the background cleanup routine
properly removes expired items from the cache at regular intervals.
*/
func TestCacheExpiredItemsCleanup(t *testing.T) {
	cache := NewCache()
	defer cache.Close()

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i, 2*time.Second)
	}	
	if len(cache.Keys()) != 10 {
		t.Errorf("Expected 10 items, got %d", len(cache.Keys()))
	}
	time.Sleep(3 * time.Second)
	
	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 items after cleanup, got %d: %v", len(keys), keys)
	}
}

/*
This function tests that an item's TTL can be extended
by setting it again with a new TTL before it expires.
*/
func TestCacheExtendTTL(t *testing.T) {
	cache := NewCache()
	defer cache.Close()

	cache.Set("key", "initial", 3*time.Second)

	time.Sleep(time.Second)
	
	cache.Set("key", "extended", 4*time.Second)
	
	time.Sleep(3*time.Second)
	
	value, found := cache.Get("key")
	if !found {
		t.Error("Expected item to still exist after TTL extension")
	}
	if value != "extended" {
		t.Errorf("Expected value 'extended', got %v", value)
	}
	
	time.Sleep(2*time.Second)
	
	_, found = cache.Get("key")
	if found {
		t.Error("Expected item to be gone after extended TTL expired")
	}
}

/*
This function tests that expired items are cleaned up by
the background process without requiring Get calls to trigger cleanup.
*/
func TestBackgroundCleanupOnly(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("bgkey%d", i)
		cache.Set(key, i, 2*time.Second)
	}
	
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("permkey%d", i)
		cache.Set(key, i)
	}
	
	if len(cache.Keys()) != 10 {
		t.Errorf("Expected 10 items initially, got %d", len(cache.Keys()))
	}
	
	time.Sleep(3 * time.Second)
	
	keys := cache.Keys()
	if len(keys) != 5 {
		t.Errorf("Expected 5 permanent items after background cleanup, got %d", len(keys))
	}
	
	for _, key := range keys {
		if key[:4] != "perm" {
			t.Errorf("Expected only permanent keys to remain, found: %s", key)
		}
	}
}

/*
This function tests that expired items are properly identified
and removed when Get is called, even if the background cleanup hasn't run yet
*/
func TestCacheExpiredItemCleanupOnGet(t *testing.T) {
	cache := NewCache()
	defer cache.Close()
	
	cache.Set("expired-item", "value", 1 * time.Second)
	
	time.Sleep(2 * time.Second)
	
	_, found := cache.Get("expired-item")
	if found {
		t.Error("Expected item to be reported as expired when Get is called")
	}
	
	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected expired item to be removed, found keys: %v", keys)
	}
}
