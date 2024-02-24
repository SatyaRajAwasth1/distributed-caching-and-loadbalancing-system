package main

import (
	"distributed-caching-and-loadbalancing-system/caching/server"
	"flag"
	"fmt"
)

func main() {
	fmt.Println("\n````````````````````````````````````````````````````````````````")

	// Handle CLI operations
	//cache.HandleCli()

	makeMaster := flag.Bool("master", false, "Run as master")
	port := flag.String("port", "8080", "Port to run on")
	flag.Parse()

	if *makeMaster {
		server.RunAsMaster(*port)
	} else {
		server.RunAsSlave(*port)
	}

	fmt.Println("````````````````````````````````````````````````````````````````")
}
