package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	masterNode = "127.0.0.1:8080" // Address of the master node
	slaveNodes = []string{
		"127.0.0.1:9090", // Address of slave node 1
		"127.0.0.1:9091", // Address of slave node 2
		"127.0.0.1:9092", // Address of slave node 3
		// Add more slave nodes as needed
	}
	slaveIndex = 0 // Index to keep track of the last used slave node
	slaveMutex sync.Mutex
)

func main() {
	// Start load balancer
	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("Error starting load balancer:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Load balancer started on port 8888")

	// Accept incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read command from client
	reader := bufio.NewReader(conn)
	command, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading command:", err)
		return
	}

	// Process command
	response := processCommand(command)

	// Send response to client
	conn.Write([]byte(response + "\n"))
}

func processCommand(command string) string {
	// Parse command
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return "Invalid command"
	}
	cmd := strings.ToUpper(parts[0])

	// Route command based on command type
	switch cmd {
	case "SET", "DEL", "FLUSHALL":
		return "Write operation routed to master node"
	case "GET":
		// Round-robin load balancing for GET commands
		slaveAddr := getNextSlave()
		return fmt.Sprintf("Read operation routed to slave node: %s", slaveAddr)
	default:
		return "Invalid command"
	}
}

func getNextSlave() string {
	slaveMutex.Lock()
	defer slaveMutex.Unlock()
	addr := slaveNodes[slaveIndex]
	slaveIndex = (slaveIndex + 1) % len(slaveNodes)
	return addr
}
