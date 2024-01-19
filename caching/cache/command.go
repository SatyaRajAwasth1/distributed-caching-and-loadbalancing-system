package cache

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ANSI escape codes for text colors
const (
	RedColor   = "\033[1;31m"
	ResetColor = "\033[0m"
)

// Command represents a cache operation command
type Command string

const (
	CMDSet      Command = "SET"
	CMDGet      Command = "GET"
	CMDDel      Command = "DEL"
	CMDFlushAll Command = "FLUSHALL"
)

// MessageSet represents a SET command message
type MessageSet struct {
	Key   string
	Value string
	TTL   time.Duration
}

// MessageGet represents a GET command message
type MessageGet struct {
	Key string
}

func HandleCli() {
	// Create a new cache
	aofFilePathPtr := "aof.log"
	cacheInstance := NewCache(aofFilePathPtr)
	defer cacheInstance.ClosePersist()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nCache:")
		cacheInstance.PrintCache()

		fmt.Print("Enter command (type 'help' for command guide): ")
		scanner.Scan()
		commandInput := scanner.Text()

		parts := strings.Fields(commandInput)
		if len(parts) == 0 {
			fmt.Println(RedColor + "Error: Empty command. Type 'help' for command guide." + ResetColor)
			continue
		}

		switch strings.ToUpper(parts[0]) {
		case string(CMDSet):
			handleSetCommand(cacheInstance, parts[1:])

		case string(CMDGet):
			handleGetCommand(cacheInstance, parts[1:])

		case string(CMDDel):
			handleDeleteCommand(cacheInstance, parts[1:])

		case string(CMDFlushAll):
			handleFlushAllCommand(cacheInstance)

		case "EXIT":
			fmt.Println("Exiting the application.")
			os.Exit(0)

		case "HELP":
			displayCommandGuide()

		default:
			fmt.Println(RedColor + "Error: Invalid command. Type 'help' for command guide." + ResetColor)
		}
		cacheInstance.PrintCache()
	}
}

func handleSetCommand(c *Cache, args []string) {
	if len(args) != 3 {
		fmt.Println(RedColor + "Error: Invalid SET command. Usage: SET <key> <value> <TTL>" + ResetColor)
		return
	}

	key := args[0]
	value := args[1]
	ttl, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("Error: Time to Live (TTL) must be a numeric value.")
		return
	}

	message := MessageSet{
		Key:   key,
		Value: value,
		TTL:   time.Duration(ttl) * time.Second,
	}

	setCache(c, message.Key, message.Value, message.TTL)
}

func handleGetCommand(c *Cache, args []string) {
	if len(args) != 1 {
		fmt.Println("Error: Invalid GET command. Usage: GET <key>")
		return
	}

	key := args[0]

	message := MessageGet{
		Key: key,
	}

	getCache(c, message.Key)
}

func handleDeleteCommand(c *Cache, args []string) {
	if len(args) != 1 {
		fmt.Println("Error: Invalid DEL command. Usage: DEL <key>")
		return
	}

	key := args[0]

	deleteCache(c, key)
}

func handleFlushAllCommand(c *Cache) {
	flushAll(c)
}

func displayCommandGuide() {
	fmt.Println("Command Guide:")
	fmt.Println("SET: Set a cache entry")
	fmt.Println("GET: Get the value of a cache entry")
	fmt.Println("DEL: Delete a cache entry")
	fmt.Println("FLUSHALL: Flush all cache entries")
	fmt.Println("EXIT: Exit the application")
	fmt.Println("HELP: Display this command guide")
	fmt.Println()
}

func setCache(c *Cache, key string, value string, duration time.Duration) {
	if key == "" || duration == 0 {
		fmt.Println("Error: Key, value, and duration are required for 'SET' command.")
		os.Exit(1)
	}
	err := c.Set(key, []byte(value), duration)
	if err != nil {
		fmt.Println("Error setting cache:", err)
	}
}

func getCache(c *Cache, key string) {
	if key == "" {
		fmt.Println("Error: Key is required for 'GET' command.")
		os.Exit(1)
	}
	data, err := c.Get(key)
	if err != nil {
		fmt.Println("Error getting cache:", err)
	} else {
		fmt.Println("Cache value:", string(data))
	}
}

func deleteCache(c *Cache, key string) {
	if key == "" {
		fmt.Println("Error: Key is required for 'DEL' command.")
		os.Exit(1)
	}
	err := c.Delete(key)
	if err != nil {
		fmt.Println("Error deleting cache:", err)
	} else {
		fmt.Println("Cache entry deleted.")
	}
}

func flushAll(c *Cache) {
	err := c.ResetCache()
	if err != nil {
		fmt.Println("Error flushing cache:", err)
	} else {
		fmt.Println("Cache flushed.")
	}
}
