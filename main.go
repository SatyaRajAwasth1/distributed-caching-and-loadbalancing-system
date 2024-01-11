package main

import (
	cache2 "distributed-caching-and-loadbalancing-system/caching/cache"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Its just a start of distributed cache and load balancing system")

	cache := cache2.New()

	// Insert elements into the cache
	cache.Set("key1", []byte("value1"), 500*time.Second)
	cache.Set("key2", []byte("value2"), 500*time.Second)

	// Display the cache content
	fmt.Println("Cache after insertion:")
	//cache.queue.Display()

	// Access an element to promote it
	cache.Get("key1")

	// Insert another element, causing eviction
	cache.Set("key3", []byte("value3"), 5*time.Second)

	// Display the updated cache content
	fmt.Println("Cache after access and insertion (eviction may occur):")
	//cache.queue.Display()
}
