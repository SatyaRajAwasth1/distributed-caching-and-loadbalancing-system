package cache

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
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

	printASCIIArt()

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
			handleSetCommand(parts[1:])

		case string(CMDGet):
			handleGetCommand(parts[1:])

		case string(CMDDel):
			handleDeleteCommand(parts[1:])

		case string(CMDFlushAll):
			handleFlushAllCommand()

		case string(CMDShowAll):
			handleShowAllCommand()

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

func printASCIIArt() {
	// Read contents of banner.txt
	banner, err := os.ReadFile("banner.txt")
	if err != nil {
		fmt.Println("Error reading banner.txt:", err)
		os.Exit(1)
	}

	// Print ASCII art
	fmt.Println(string(banner))
}

func handleSetCommand(args []string) {
	if len(args) != 3 {
		fmt.Println(RedColor + "Error: Invalid SET command. Usage: SET <key> <value> <TTL>" + ResetColor)
		return
	}

	key := args[0]
	value := args[1]
	ttl := args[2]

	message := MessageSet{
		Key:   key,
		Value: []byte(value),
		TTL:   parseTTL(ttl),
	}

	resp, err := sendSetRequest(message)
	if err != nil {
		fmt.Println(RedColor+"Error setting cache:", err, ResetColor)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var responseData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			fmt.Println(RedColor+"Error decoding response:", err, ResetColor)
			return
		}

		printCacheData(responseData)
	} else {
		fmt.Println(RedColor+"Error:", resp.Status+ResetColor)
	}
}

func handleGetCommand(args []string) {
	if len(args) != 1 {
		fmt.Println(RedColor + "Error: Invalid GET command. Usage: GET <key>" + ResetColor)
		return
	}

	key := args[0]

	message := MessageGet{
		Key: key,
	}

	resp, err := sendGetRequest(message)
	if err != nil {
		fmt.Println(RedColor+"Error getting cache:", err, ResetColor)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var responseData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			fmt.Println(RedColor+"Error decoding response:", err, ResetColor)
			return
		}

		printCacheData(responseData)
	} else {
		fmt.Println(RedColor+"Error:", resp.Status+ResetColor)
	}
}

func handleDeleteCommand(args []string) {
	if len(args) != 1 {
		fmt.Println(RedColor + "Error: Invalid DEL command. Usage: DEL <key>" + ResetColor)
		return
	}

	key := args[0]

	resp, err := sendDeleteRequest(key)
	if err != nil {
		fmt.Println(RedColor+"Error deleting cache:", err, ResetColor)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Cache entry deleted.")
	} else {
		fmt.Println(RedColor+"Error:", resp.Status+ResetColor)
	}
}

func handleFlushAllCommand() {
	resp, err := sendFlushAllRequest()
	if err != nil {
		fmt.Println(RedColor+"Error flushing cache:", err, ResetColor)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Cache flushed.")
	} else {
		fmt.Println(RedColor+"Error:", resp.Status+ResetColor)
	}
}

func handleShowAllCommand() {
	resp, err := sendShowAllRequest()
	if err != nil {
		fmt.Println(RedColor+"Error getting cache data:", err, ResetColor)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var responseData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			fmt.Println(RedColor+"Error decoding response:", err, ResetColor)
			return
		}

		println("Response Body: ", resp.Body)
		printCacheData(responseData)
	} else {
		fmt.Println(RedColor+"Error:", resp.Status+ResetColor)
	}
}

func parseTTL(ttl string) time.Duration {
	duration, err := strconv.Atoi(ttl)
	if err != nil {
		fmt.Println(RedColor + "Error: Time to Live (TTL) must be a numeric value." + ResetColor)
		os.Exit(1)
	}
	return time.Duration(duration) * time.Second
}

func printCacheData(data map[string]interface{}) {
	fmt.Println("Cache Data:")
	for key, value := range data {
		fmt.Printf("%s: %s\n", key, value)
	}
}

func sendSetRequest(message MessageSet) (*http.Response, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8888/cache/set", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendGetRequest(message MessageGet) (*http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8888/cache/get?key=%s", message.Key))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendDeleteRequest(key string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8888/cache/delete?key=%s", key), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendFlushAllRequest() (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:8888/cache/reset", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendShowAllRequest() (*http.Response, error) {
	resp, err := http.Get("http://localhost:8888/cache/getAll")
	if err != nil {
		return nil, err
	}

	return resp, nil
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
