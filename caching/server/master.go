package server

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"encoding/json"
	"fmt"
	"net"
)

func RunAsMaster(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Running as master on port", port)
	aofUrl, _ := getAofFileLocation("config.yaml")
	cacheInstance := cache.NewCache()
	cacheInstance.ReplayAOF(aofUrl)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleSlaveConnection(conn)
	}
}

func handleSlaveConnection(conn net.Conn) {
	defer conn.Close()

	// Read slave information from the connection
	slaveInfo := NodeInfo{}
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&slaveInfo)
	if err != nil {
		fmt.Println("Error reading slave information:", err)
		return
	}

	fmt.Println("Slave connected:", slaveInfo)
}
