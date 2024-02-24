package server

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"encoding/json"
	"fmt"
	"net"
)

// RunAsMaster starts the master node.
func RunAsMaster(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Running as master on port", port)
	aofUrl := "tmp/aof.log" //getAofFileLocation("config.yaml")
	cacheInstance := cache.NewCache()
	fmt.Println("AOF URL:", aofUrl)
	cacheInstance.ReplayAOF(aofUrl)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleSlaveConnection(conn, cacheInstance)
	}
}

// handleSlaveConnection handles connections from slave nodes.
func handleSlaveConnection(conn net.Conn, cacheInstance *cache.Cache) {
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

	// Send cache data to the slave for replication
	cacheData := cacheInstance.GetCacheData()
	encoder := json.NewEncoder(conn)
	err = encoder.Encode(cacheData)
	if err != nil {
		fmt.Println("Error sending cache data to slave:", err)
		return
	}
}
