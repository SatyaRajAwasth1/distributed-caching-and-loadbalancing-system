package server

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func RunAsSlave(port string) {
	// Read configuration file
	masterAddr, masterPort, err := readMasterNodeConfigs("config.yml")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	// Connect to master
	conn, err := net.Dial("tcp", masterAddr+":"+masterPort)
	if err != nil {
		fmt.Println("Error connecting to master:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to master at", masterAddr+":"+masterPort)

	// Prepare slave information
	slaveInfo := NodeInfo{
		NodeId:     generateUniqueId(),
		NodeIpAddr: getOutboundIP(),
		Port:       port,
	}

	// Send JSON-encoded slave information to master
	encoder := json.NewEncoder(conn)
	err = encoder.Encode(slaveInfo)
	if err != nil {
		fmt.Println("Error sending slave information to master:", err)
		return
	}

	// Receive cache data from master
	var cacheData map[string][]byte
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&cacheData)
	if err != nil {
		fmt.Println("Error receiving cache data from master:", err)
		return
	}

	// Initialize local cache with received data
	cacheInstance := cache.NewCache()
	cacheInstance.SetCacheData(cacheData)
}

func generateUniqueId() int {
	// You can implement your unique ID generation logic here
	// For simplicity, let's use the current timestamp as the ID
	return int(time.Now().Unix())
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Error getting outbound IP:", err)
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
