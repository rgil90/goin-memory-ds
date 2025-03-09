# Go In-Memory Cache

A lightweight, thread-safe in-memory caching system written in Go that provides Redis-like functionality for storing and retrieving data with optional expiration times.

## What This Program Does

This cache implementation provides a simple way to store and retrieve data in memory with these key features:

- **Key-Value Storage**: Store any type of data using string keys
- **Time-To-Live (TTL)**: Set optional expiration times for cached items
- **Automatic Cleanup**: Background process removes expired items
- **Thread Safety**: Safe for concurrent access from multiple goroutines

## How It's Similar to Redis

This implementation is similar to Redis in several ways:

- **Simple Key-Value API**: Uses familiar Set/Get/Delete operations
- **Optional TTL**: Keys can expire automatically after a specified time
- **Persistence of Non-Expiring Keys**: Keys without TTL remain until explicitly removed
- **Atomic Operations**: Operations are thread-safe, similar to Redis's single-threaded model

While Redis offers many more features (like persistence to disk, pub/sub, different data structures), this implementation focuses on providing the core in-memory caching functionality that many applications need.

## Thread Safety & Concurrency

### How It's Thread-Safe

The cache uses Go's synchronization primitives to ensure thread safety:

- **Read-Write Mutex (`sync.RWMutex`)**: Protects access to the cache data
- **Read Locks**: For operations that don't modify the cache (like Get)
- **Write Locks**: For operations that modify the cache (like Set, Delete)

This approach allows multiple readers to access the cache simultaneously, but ensures exclusive access during writes to prevent data corruption.

### How It Handles Concurrency

The implementation uses several strategies to handle concurrent access efficiently:

1. **Background Cleanup**: A dedicated goroutine periodically scans for and removes expired items
2. **Lazy Deletion**: The Get method also checks for expiration, removing items if they've expired
3. **Fine-Grained Locking**: Locks are held for the minimum duration necessary
4. **Safe Map Operations**: Map-related operations are always protected by appropriate locks

This design ensures that the cache can handle high-concurrency workloads without performance degradation or data corruption.

## Usage Example

```go
import (
    "github.com/rgil90/goin-memory-ds/modules"
    "time"
    "fmt"
)

func main() {
    // Create a new cache
    cache := modules.NewCache()
    defer cache.Close() // Ensure proper cleanup

    // Set values (with and without TTL)
    cache.Set("permanent-key", "This value never expires")
    cache.Set("temporary-key", "Expires in 5 minutes", 5*time.Minute)
    
    // Retrieve values
    value, found := cache.Get("permanent-key")
    if found {
        fmt.Println(value) // "This value never expires"
    }
    
    // Delete values
    cache.Delete("temporary-key")
    
    // List all keys
    keys := cache.Keys()
    fmt.Println(keys) // ["permanent-key"]
    
    // Clear the cache
    cache.Clear()
}
```

## Running Tests

Tests are written using the testify assertion library for better readability and error messages:

```
go test -v ./modules
```

## Future Improvements

Planned enhancements to make this cache even more powerful:

1. **LRU Key Eviction**: Automatically remove least-recently-used items when memory usage exceeds a threshold
   - This would complement the TTL-based expiration by ensuring memory usage remains bounded
   - Would include configurable maximum item count or memory usage limits

2. **Additional Eviction Policies**:
   - LFU (Least Frequently Used)
   - FIFO (First-In-First-Out)
   - Random Eviction

3. **Statistics and Monitoring**:
   - Hit/miss ratio tracking
   - Memory usage statistics
   - Expiration/eviction counters

4. **Sharding**: Divide the cache into multiple partitions to reduce lock contention

5. **Disk Persistence**: Optional saving/loading of cache to/from disk
