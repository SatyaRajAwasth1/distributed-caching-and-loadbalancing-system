package main

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"distributed-caching-and-loadbalancing-system/caching/server"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Its just a start of distributed cache and load balancing system")

	serverOperations := server.Operations{
		ListenAddress: "127.0.0.1:8080",
		IsLeader:      true,
	}

	newServer := server.NewServer(serverOperations, cache.New())
	err := newServer.Start()
	if err != nil {
		log.Printf("Error starting server: %s", err)
	}
}
