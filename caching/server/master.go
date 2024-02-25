package server

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	startTime        = time.Now()
	slaveConnections = make(map[string]net.Conn)
)

// Info represents information about the server.
type Info struct {
	Status          string    `json:"status"`
	StartTime       time.Time `json:"start_time"`
	ConnectedSlaves []string  `json:"connected_slaves"`
	NumberOfSlaves  int       `json:"number_of_slaves"`
}

// RunAsMaster starts the master node.
func RunAsMaster(port string) {
	aofUrl := "tmp/aof.log" //getAofFileLocation("config.yaml")
	cacheInstance := cache.NewCache()
	fmt.Println("Master listening to slaves at port: 8080")
	fmt.Println("Listening to clients at port ", port)
	fmt.Println("AOF URL:", aofUrl)
	cacheInstance.ReplayAOF(aofUrl)

	// Expose HTTP endpoints for cache operations
	http.HandleFunc("/cache/set", handleSetCache(cacheInstance))
	http.HandleFunc("/cache/delete", handleDeleteCache(cacheInstance))
	http.HandleFunc("/cache/reset", handleResetCache(cacheInstance))
	http.HandleFunc("/server/info", handleServerInfo)

	// Start HTTP server for clients
	go func() {
		fmt.Println("HTTP server for clients running on port ", port)
		log.Fatal(http.ListenAndServe(":"+port, logRequest(http.DefaultServeMux)))
	}()
	println("Out of the routine")
	// Accept slave connection requests and handle them
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle slave connection
		go handleSlaveConnection(conn, cacheInstance)
	}
}

// handleSlaveConnection handles connections from slave nodes.
func handleSlaveConnection(conn net.Conn, cacheInstance *cache.Cache) {
	defer conn.Close()

	log.Printf("Handling slave request....")
	// Read slave information from the connection
	slaveInfo := NodeInfo{}
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&slaveInfo); err != nil {
		log.Println("Error reading slave information:", err)
		return
	}
	log.Printf("Slave connected: %v\n", slaveInfo)

	// Add slave connection to the map
	slaveConnections[string(rune(slaveInfo.NodeId))] = conn

	// Send cache data to the slave for replication
	if err := sendCacheDataToSlave(conn, cacheInstance); err != nil {
		log.Println("Error sending cache data to slave:", err)
		return
	}

	// Log when slave disconnects
	defer func() {
		delete(slaveConnections, strconv.Itoa(slaveInfo.NodeId))
		log.Printf("Slave disconnected: %v\n", slaveInfo)
	}()
}

// sendCacheDataToSlave sends cache data to the slave for replication.
func sendCacheDataToSlave(conn net.Conn, cacheInstance *cache.Cache) error {
	// Obtain cache data
	cacheData := cacheInstance.GetCacheData()

	println("Cache Data: ", cacheData)
	// Encode and send cache data to the slave
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cacheData); err != nil {
		return fmt.Errorf("error sending cache data to slave: %v", err)
	}

	return nil
}

// handleSetCache handles the cache set operation on the master node.
func handleSetCache(cacheInstance *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Set cache request received: %s %s", r.Method, r.URL.Path)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode JSON request body
		var request struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			TTL   string `json:"ttl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
			return
		}

		// Perform cache set operation with TTL
		if request.TTL != "" {
			// Parse TTL duration
			duration, err := time.ParseDuration(request.TTL)
			if err != nil {
				http.Error(w, "Invalid TTL duration", http.StatusBadRequest)
				return
			}

			// Set cache with TTL
			err = cacheInstance.Set(request.Key, []byte(request.Value), duration)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotAcceptable)
				return
			}
		}

		// Respond with success message
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Cache set successful\n")
	}
}

// handleDeleteCache handles the cache delete operation on the master node.
func handleDeleteCache(cacheInstance *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP request received: %s %s", r.Method, r.URL.Path)
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract key from the query string
		key := r.URL.Query().Get("key")

		if key == "" {
			http.Error(w, "Missing key parameter", http.StatusBadRequest)
			return
		}

		// Perform cache delete operation
		err := cacheInstance.Delete(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with success message
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Cache delete successful\n")
	}
}

// handleResetCache handles the cache reset operation on the master node.
func handleResetCache(cacheInstance *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP request received: %s %s", r.Method, r.URL.Path)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Perform cache reset operation
		err := cacheInstance.ResetCache()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with success message
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Cache reset successful\n")
	}
}

// handleServerInfo handles requests for server information.
func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP request received: %s %s", r.Method, r.URL.Path)
	info := Info{
		Status:          "Running",
		StartTime:       startTime,
		ConnectedSlaves: getConnectedSlaveIDs(),
		NumberOfSlaves:  len(slaveConnections),
	}

	// Encode server information as JSON and write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// getConnectedSlaveIDs returns the IDs of connected slaves.
func getConnectedSlaveIDs() []string {
	var ids []string
	for id := range slaveConnections {
		ids = append(ids, id)
	}
	return ids
}

// logRequest is a middleware function to log HTTP requests.
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP request received: %s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}
