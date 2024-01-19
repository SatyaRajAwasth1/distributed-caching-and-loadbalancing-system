package main

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"fmt"
)

func main() {
	fmt.Println("\n````````````````````````````````````````````````````````````````")

	// Handle CLI operations
	cache.HandleCli()

	fmt.Println("````````````````````````````````````````````````````````````````")
}
