package server

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func RunAsSlave(port string) {
	// Read configuration file
	masterAddr, masterPort, err := readMasterNodeConfigs("config.yml")
	if err != nil {
		log.Println("Error reading config file:", err)
		return
	}

	// Retry connecting to master with backoff
	for {
		conn, err := net.Dial("tcp", masterAddr+":"+masterPort)
		if err != nil {
			log.Println("Error connecting to master:", err)
			time.Sleep(5 * time.Second) // Retry after 5 seconds
			continue
		}
		defer conn.Close()

		log.Println("Connected to master at", masterAddr+":"+masterPort)

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
			log.Println("Error sending slave information to master:", err)
			return
		}

		// Receive cache data from master
		var cacheData map[string][]byte
		decoder := json.NewDecoder(conn)
		println("Response: ", decoder)
		err = decoder.Decode(&cacheData)
		if err != nil {
			log.Println("Error receiving cache data from master:", err)
		}

		// Initialize local cache with received data
		cacheInstance := cache.NewCache()
		cacheInstance.SetCacheData(cacheData)

		// Expose HTTP endpoints for read operations
		http.HandleFunc("/cache/get", handleGetCache(cacheInstance))
		http.HandleFunc("/cache/getAll", handleGetAllCacheData(cacheInstance))

		// Start HTTP server
		log.Fatal(http.ListenAndServe(":"+port, nil))

		// Exit the retry loop if connection and initialization are successful
		break
	}
}

// handleGetCache handles the cache get operation on the slave node.
func handleGetCache(cacheInstance *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received GET request for cache")
		defer log.Println("GET request processed")

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract key from the request parameters
		keys, ok := r.URL.Query()["key"]
		if !ok || len(keys) < 1 {
			http.Error(w, "Missing key parameter", http.StatusBadRequest)
			return
		}
		key := keys[0]

		// Perform cache get operation
		value, err := cacheInstance.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		println("Response Received: ", string(value))
		// Respond with cache value
		w.WriteHeader(http.StatusOK)
		w.Write(value)
	}
}

// handleGetAllCacheData handles the request for retrieving all cache data (map and queue info).
func handleGetAllCacheData(cacheInstance *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received GET request for all cache data")
		defer log.Println("GET request for all cache data processed")

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get cache map data
		cacheMap := cacheInstance.GetCacheData()

		// Prepare queue data
		queueData := make([]interface{}, 0)
		currentNode := cacheInstance.Queue.Head
		for currentNode != nil {
			queueData = append(queueData, currentNode.Data)
			currentNode = currentNode.Next
		}

		// Prepare response object
		responseData := struct {
			CacheMap map[string][]byte `json:"cache_map"`
			Queue    []interface{}     `json:"queue"`
		}{
			CacheMap: cacheMap,
			Queue:    queueData,
		}

		// Encode and send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responseData); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func generateUniqueId() int {
	// use the current timestamp as the ID
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
