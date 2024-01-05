package server

import (
	"distributed-caching-and-loadbalancing-system/caching/cache"
	"fmt"
	"log"
	"net"
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
		fmt.Println(message)
	}
}
