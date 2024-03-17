package main

import (
	"bufio"
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	masterNode  string   // Address of the master node
	slaveNodes  []string // Addresses of slave nodes
	slaveIndex  = 0      // Index to keep track of the last used slave node
	slaveMutex  sync.Mutex
	configMutex sync.Mutex
)

// Config struct to hold master and slave configurations
type Config struct {
	Cache struct {
		Master struct {
			Address string `yaml:"address"`
			Port    string `yaml:"port"`
		} `yaml:"master"`
		Slaves     []string `yaml:"slaves"`
		AofFileUrl string   `yaml:"aof"`
	} `yaml:"cache"`
}

func main() {
	// Read configuration from YAML file
	config, err := readConfig("config.yml")
	if err != nil {
		log.Println("Error reading config file:", err)
		return
	}

	masterNode = "127.0.0.1:8081" //config.Cache.Master.Address + ":" + config.Cache.Master.Port not using since thats the tcp listening to slaves
	slaveNodes = config.Cache.Slaves

	// Start load balancer
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Println("Error starting load balancer:", err)
		return
	}
	defer ln.Close()
	log.Println("Load balancer started on port 8888")

	// Handle incoming requests
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	// Read HTTP request
	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Println("Error reading HTTP request:", err)
		return
	}

	// Log the request details
	log.Printf("Received %s request for %s", request.Method, request.URL.Path)

	// Route request to appropriate node
	nodeAddr := routeRequest(request)

	// Forward request to node
	forwardRequest(conn, nodeAddr, request)

	// Log the completion of handling the request
	log.Println("Request handling completed")
}

func routeRequest(req *http.Request) string {
	// Extract request path
	path := req.URL.Path

	// Route based on request path
	if strings.HasPrefix(path, "/server/info") || req.Method == http.MethodPost {
		// GET request for server info or POST request, route to master node
		return masterNode
	}

	// Round-robin load balancing for other requests
	slaveAddr := getNextSlave()
	return slaveAddr
}

func getNextSlave() string {
	slaveMutex.Lock()
	defer slaveMutex.Unlock()
	addr := slaveNodes[slaveIndex]
	slaveIndex = (slaveIndex + 1) % len(slaveNodes)
	return addr
}

func forwardRequest(conn net.Conn, nodeAddr string, req *http.Request) {
	// Connect to node
	nodeConn, err := net.Dial("tcp", nodeAddr)
	if err != nil {
		log.Println("Error connecting to node:", err)
		return
	}
	defer nodeConn.Close()

	// Log the request forwarding details
	parts := strings.Split(nodeAddr, ":")
	serverIP := parts[0]
	serverPort := parts[1]
	log.Printf("Forwarding request to server %s:%s\n", serverIP, serverPort)

	// Forward request to node
	err = sendRequest(nodeConn, req)
	if err != nil {
		log.Println("Error forwarding request to node:", err)
		return
	}

	// Forward response from node to client
	err = forwardResponse(conn, nodeConn)
	if err != nil {
		log.Println("Error forwarding response from node to client:", err)
		return
	}
}

func sendRequest(serverConn net.Conn, req *http.Request) error {
	// Write request to server connection
	err := req.Write(serverConn)
	if err != nil {
		return err
	}
	return nil
}

func forwardResponse(clientConn net.Conn, serverConn net.Conn) error {
	// Read response from server
	resp, err := http.ReadResponse(bufio.NewReader(serverConn), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write response to client connection
	var buf bytes.Buffer
	resp.Write(&buf)
	_, err = io.Copy(clientConn, &buf)
	if err != nil {
		return err
	}

	// Log the response details
	log.Println("Response forwarded to client")

	return nil
}

func readConfig(filename string) (*Config, error) {
	// Lock to prevent concurrent access while reading configuration
	configMutex.Lock()
	defer configMutex.Unlock()

	// Read YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse YAML data into Config struct
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
