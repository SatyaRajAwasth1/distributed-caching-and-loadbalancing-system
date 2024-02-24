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
	GreenColor = "\033[1;32m"
	BlueColor  = "\033[1;34m"
	ResetColor = "\033[0m"
)

// Command represents a cache operation command
type Command string

const (
	CMDSet      Command = "SET"
	CMDGet      Command = "GET"
	CMDDel      Command = "DEL"
	CMDFlushAll Command = "FLUSHALL"
	CMDShowAll  Command = "SHOWALL"
	CMDExit     Command = "EXIT"
	CMDHelp     Command = "HELP"
)

// MessageSet represents a SET command message
type MessageSet struct {
	Key   string
	Value []byte
	TTL   time.Duration
}

// MessageGet represents a GET command message
type MessageGet struct {
	Key string
}

func HandleCli() {
	// Create a new cache
	cacheInstance := NewCache()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n" + BlueColor + "Enter command (type '" + GreenColor + "help" + BlueColor + "' for command guide)> " + ResetColor)
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
			cacheInstance.PrintCache()

		case string(CMDGet):
			handleGetCommand(cacheInstance, parts[1:])

		case string(CMDDel):
			handleDeleteCommand(cacheInstance, parts[1:])
			cacheInstance.PrintCache()

		case string(CMDFlushAll):
			handleFlushAllCommand(cacheInstance)
			cacheInstance.PrintCache()

		case string(CMDShowAll):
			cacheInstance.PrintCache()

		case string(CMDExit):
			fmt.Println("Exiting the application.")
			os.Exit(0)

		case string(CMDHelp):
			displayCommandGuide()

		default:
			fmt.Println(RedColor + "Error: Invalid command. Type 'help' for command guide." + ResetColor)
		}
	}
}

func handleSetCommand(c *Cache, args []string) {
	if len(args) != 3 {
		fmt.Println(RedColor + "Error: Invalid SET command. Usage: SET <key> <value> <TTL>" + ResetColor)
		return
	}

	key := args[0]
	value := []byte(args[1])
	ttl, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Println(RedColor + "Error: Time to Live (TTL) must be a numeric value." + ResetColor)
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
		fmt.Println(RedColor + "Error: Invalid GET command. Usage: GET <key>" + ResetColor)
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
		fmt.Println(RedColor + "Error: Invalid DEL command. Usage: DEL <key>" + ResetColor)
		return
	}

	key := args[0]

	deleteCache(c, key)
}

func handleFlushAllCommand(c *Cache) {
	flushAll(c)
}

func displayCommandGuide() {
	fmt.Println("🚀 " + BlueColor + "Command Guide" + ResetColor + " 🚀")
	fmt.Println("----------------------------------")
	fmt.Printf(" %-14s | %s\n", GreenColor+"SET"+ResetColor, "Set a cache entry")
	fmt.Printf(" %-14s | %s\n", GreenColor+"   Usage:"+ResetColor, "SET <key> <value> <TTL>")
	fmt.Printf(" %-14s | %s\n", GreenColor+"GET"+ResetColor, "Get the value of a cache entry")
	fmt.Printf(" %-14s | %s\n", GreenColor+"   Usage:"+ResetColor, "GET <key>")
	fmt.Printf(" %-14s | %s\n", GreenColor+"DEL"+ResetColor, "Delete a cache entry")
	fmt.Printf(" %-14s | %s\n", GreenColor+"   Usage:"+ResetColor, "DEL <key>")
	fmt.Printf(" %-14s | %s\n", GreenColor+"FLUSHALL"+ResetColor, "Flush all cache entries")
	fmt.Printf(" %-14s | %s\n", GreenColor+"EXIT"+ResetColor, "Exit the application")
	fmt.Printf(" %-14s | %s\n", GreenColor+"HELP"+ResetColor, "Display this command guide")
	fmt.Println("----------------------------------")
	fmt.Println()
}

func setCache(c *Cache, key string, value []byte, duration time.Duration) {
	if key == "" || duration == 0 {
		fmt.Println(RedColor + "Error: Key, value, and duration are required for 'SET' command." + ResetColor)
		os.Exit(1)
	}
	err := c.Set(key, []byte(value), duration)
	if err != nil {
		fmt.Println(RedColor+"Error setting cache:", err, ResetColor)
	}
}

func getCache(c *Cache, key string) {
	if key == "" {
		fmt.Println(RedColor + "Error: Key is required for 'GET' command." + ResetColor)
		os.Exit(1)
	}
	data, err := c.Get(key)
	if err != nil {
		fmt.Println(RedColor+"Error getting cache:", err, ResetColor)
	} else {
		fmt.Println("Cache value:", string(data))
	}
}

func deleteCache(c *Cache, key string) {
	if key == "" {
		fmt.Println(RedColor + "Error: Key is required for 'DEL' command." + ResetColor)
	}
	err := c.Delete(key)
	if err != nil {
		fmt.Println(RedColor+"Error deleting cache:", err, ResetColor)
	} else {
		fmt.Println("Cache entry deleted.")
	}
}

func flushAll(c *Cache) {
	err := c.ResetCache()
	if err != nil {
		fmt.Println(RedColor+"Error flushing cache:", err, ResetColor)
	} else {
		fmt.Println("Cache flushed.")
	}
}
