package main

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"distributed-caching-and-loadbalancing-system/caching/server"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	fmt.Println("Its just a start of distributed cache and load balancing system")

	serverOperations := server.Operations{
		ListenAddress: ":8080",
		IsLeader:      true,
	}

	go func() {
		time.Sleep(time.Second * 2)
		conn, err := net.Dial("tcp", ":8080")
		if err != nil {
			return
		}

		_, _ = conn.Write([]byte("Hello, Server From Satya"))
	}()

	newServer := server.NewServer(serverOperations, cache.New())
	err := newServer.Start()
	if err != nil {
		log.Printf("Error starting server: %s", err)
	}
}
