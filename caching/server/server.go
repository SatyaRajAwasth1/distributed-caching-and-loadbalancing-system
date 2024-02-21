package server

import (
	"context"
	"distributed-caching-and-loadbalancing-system/caching"
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type Operations struct {
	ListenAddress string
	IsLeader      bool
}

type Server struct {
	serverOperations Operations
	cache            cache.Cacher
}

func NewServer(operations Operations, cache cache.Cacher) *Server {
	return &Server{
		serverOperations: operations,
		cache:            cache,
	}
}

func (server *Server) Start() error {
	listen, err := net.Listen("tcp", server.serverOperations.ListenAddress)
	if err != nil {
		return fmt.Errorf("error listening: %s", err)
	}

	_, port, _ := net.SplitHostPort(server.serverOperations.ListenAddress)
	log.Printf("Server starting on port %s \n", port)
	for {
		connection, err := listen.Accept()
		if err != nil {
			log.Printf("Couldn't start server. Accept err: %s \n", err)
			continue
		}
		go server.handleConnection(connection)
	}
}

func (server *Server) handleConnection(connection net.Conn) {
	defer func() {
		_ = connection.Close()
	}()

	buffer := make([]byte, 2048)
	for {
		read, err := connection.Read(buffer)
		if err != nil {
			log.Printf("connection read error: %s \n", err)
			break
		}

		message := buffer[:read]
		fmt.Println(string(message))
	}
}

func (server *Server) handleCommand(connection net.Conn, rawCommand []byte) {
	var (
		rawString = string(rawCommand)
		parts     = strings.Split(rawString, " ")
	)

	if len(parts) == 0 {
		log.Println("Invalid Command")
		return
	}

	command := caching.Command(parts[0])

	switch command {
	case caching.CMDSet:
		if len(parts) != 4 {
			log.Println("Invalid SET command. The command must be of format 'SET <key> <value> <TTL>'")
			return
		}

		ttl, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Println("Time to Live (TTL) must be a numeric value.")
			return
		}

		message := caching.MessageSet{
			Key:   []byte(parts[1]),
			Value: []byte(parts[2]),
			TTL:   time.Duration(ttl) * time.Second, // Convert TTL to time.Duration
		}

		err = server.handleSetCommand(connection, &message)
		if err != nil {
			log.Println("Error handling SET command:", err)
		}

	case caching.CMDGet:
		if len(parts) != 2 {
			log.Println("Invalid GET command. The command must be of format 'GET <key>'")
			return
		}

		message := caching.MessageGet{
			Key: []byte(parts[1]),
		}

		_, err := server.handleGetCommand(connection, message)
		if err != nil {
			log.Println("Error handling GET command:", err)
		}

	default:
		log.Println("Unsupported command:", command)
	}
}

func (server *Server) handleSetCommand(connection net.Conn, message *caching.MessageSet) error {
	err := server.cache.Set(message.Key, message.Value, message.TTL)
	if err != nil {
		return err
	}
	log.Printf("SET command received. Key: %s, Value: %s, TTL: %s\n", message.Key, message.Value, message.TTL)
	go server.replicateToFollowers(context.TODO(), message)
	return nil
}

func (server *Server) handleGetCommand(connection net.Conn, message caching.MessageGet) ([]byte, error) {
	val, err := server.cache.Get(message.Key)
	if err != nil {
		return val, nil
	}
	log.Printf("GET command received. Key: %s\n", message.Key)
	return val, nil
}

func (server *Server) replicateToFollowers(context context.Context, message *caching.MessageSet) error {
	return nil
}
