package cache

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Open or create AOF file.
func openOrCreateAOFFile(aofFilePath string) (*os.File, error) {
	return os.OpenFile(aofFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (c *Cache) ReplayAOF(aofFilePath string) {
	log.Printf("Start replaying AOF commands")
	file, err := os.Open(aofFilePath)
	if err != nil {
		fmt.Println("Error opening AOF file for replay:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("Error closing AOF file.")
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 1 {
			command := fields[0]

			// Replay the command to reconstruct the cache
			switch command {
			case "SET":
				key := fields[1]
				value := []byte(fields[2])
				ttl, _ := strconv.Atoi(fields[3])
				_ = c.Set(key, value, time.Duration(ttl)) // Duration is 0 for no eviction
			case "DELETE":
				key := fields[1]
				_ = c.Delete(key)
			case "FLUSHALL":
				_ = c.ResetCache()

			}
		}
		log.Printf("Replay of AOF end")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(RedColor+"Error reading AOF file:", err, ResetColor)
	}

}

func (c *Cache) writeToAOF(command string, key string, value []byte, ttl time.Duration) {
	if c.aofFile == nil {
		return
	}

	// Format the command and write to the AOF file
	cmd := fmt.Sprintf("%s %s %s %s\n", command, key, value, ttl)
	_, err := c.aofFile.WriteString(cmd)
	if err != nil {
		fmt.Println("Error writing to AOF file:", err)
	}
}

func (c *Cache) CloseAOF() {
	if c.aofFile != nil {
		err := c.aofFile.Close()
		if err != nil {
			log.Printf("Error closing AOF file.")
			return
		}
	}
}
