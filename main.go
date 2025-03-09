package main

import (
	"fmt"
	"github.com/rgil90/goin-memory-ds/modules"
)

/*
This function is the main function that creates a new cache instance and sets some values.
It then retrieves a value, lists all keys, deletes a key, checks if a key exists after deletion,
clears the cache, and prints the size of the cache after clearing.
*/
func main() {
	// Create a new cache instance
	cache := modules.NewCache()
	
	// Set some values
	cache.Set("name", "GoCache")
	cache.Set("version", 1.0)
	cache.Set("active", true)
	
	// Retrieve a value
	if name, found := cache.Get("name"); found {
		fmt.Printf("Cache name: %v\n", name)
	}
	
	// List all keys
	keys := cache.Keys()
	fmt.Println("All cache keys:", keys)
	
	// Delete a key
	cache.Delete("active")
	
	// Check if key exists after deletion
	if _, found := cache.Get("active"); !found {
		fmt.Println("Key 'active' has been removed")
	}
	
	// Clear the cache
	cache.Clear()
	fmt.Printf("Cache size after clear: %d keys\n", len(cache.Keys()))
}
