package cache

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Open or create AOF file.
func openOrCreateAOFFile(aofFilePath string) (*os.File, error) {
	return os.OpenFile(aofFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (c *Cache) ReplayAOF(aofFilePath string) {
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
			key := fields[1]
			value := []byte(fields[2])

			// Replay the command to reconstruct the cache
			switch command {
			case "SET":
				_ = c.Set(key, value, 0) // Duration is 0 for no eviction
			case "DELETE":
				_ = c.Delete(key)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading AOF file:", err)
	}

	// Start background AOF write goroutine
	c.aofWriteTicker = time.NewTicker(1 * time.Minute)
	go c.backgroundAOFWrite()
}

func (c *Cache) backgroundAOFWrite() {
	for {
		select {
		case <-c.aofWriteTicker.C:
			c.aofMutex.Lock()
			c.writeToAOFInBackground()
			c.aofMutex.Unlock()
		}
	}
}

func (c *Cache) writeToAOFInBackground() {
	if c.aofFile == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Perform AOF write operations based on the cache contents
	for key, node := range c.CacheMap {
		value := node.data.([]byte)
		// Customize the AOF write operation based on your needs
		c.writeToAOF("SET", key, value)
	}
}

func (c *Cache) writeToAOF(command string, key string, value []byte) {
	if c.aofFile == nil {
		return
	}

	// Format the command and write to the AOF file
	cmd := fmt.Sprintf("%s %s %s\n", command, key, value)
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

func (c *Cache) ClosePersist() {
	// Stop AOF write goroutine
	c.aofWriteTicker.Stop()

	// Close AOF file
	c.CloseAOF()
}
