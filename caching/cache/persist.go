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
	c.replayingAOF = true
	aofFile, err := openOrCreateAOFFile(aofFilePath)
	if err != nil {
		fmt.Println(RedColor+"Error opening AOF file for replay:", err, ResetColor)
		return
	}
	c.aofFile = aofFile

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(RedColor + "Error closing AOF file." + ResetColor)
		}
	}(aofFile)

	scanner := bufio.NewScanner(aofFile)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 1 {
			command := fields[0]

			// Replay the command to reconstruct the cache
			switch command {
			case string(CMDSet):
				key := fields[1]
				value := []byte(fields[2])
				ttl, _ := strconv.Atoi(fields[3])
				duration := time.Duration(ttl/1000) * time.Second
				_ = c.Set(key, value, duration)
			case string(CMDDel):
				key := fields[1]
				_ = c.Delete(key)
			case string(CMDFlushAll):
				_ = c.ResetCache()

			}
		}
		log.Printf("Replay of AOF end")
	}
	c.replayingAOF = false
	if err := scanner.Err(); err != nil {
		fmt.Println(RedColor+"Error reading AOF file:", err, ResetColor)
	}

}

func (c *Cache) writeToAOF(command string, key string, value []byte, ttl time.Duration) {
	if c.aofFile == nil {
		log.Printf("Error! No AOF file.")
		return
	}

	// Format the command and write to the AOF file
	var cmd = fmt.Sprintf("%s %s %s \n", command, key, value)
	if ttl != 0 {
		cmd = fmt.Sprintf("%s %s %s %d\n", command, key, value, ttl.Milliseconds())
	}
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
