package cache

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Cache struct {
	CacheMap     map[string]*Node
	Queue        *Queue
	mutex        sync.Mutex
	aofFile      *os.File
	aofMutex     sync.Mutex
	replayingAOF bool
	Size         int // Size of the cache
	Evictions    int // Number of cache evictions
	Hits         int // Number of cache hits
	Misses       int // Number of cache misses
}

type Queue struct {
	Head *Node
	tail *Node
}

type Node struct {
	Data interface{}
	prev *Node
	Next *Node
}

// NewCache creates a new Cache with an empty map and Queue.
func NewCache() *Cache {
	return &Cache{
		CacheMap: make(map[string]*Node),
		Queue:    NewQueue(),
		mutex:    sync.Mutex{},
	}
}

// NewQueue creates a new Queue with empty head and tail.
func NewQueue() *Queue {
	return &Queue{}
}

func (c *Cache) PrintCache() {
	fmt.Println()
	fmt.Println("╔═════════════════════════════╗")
	fmt.Println("║        Cache Contents       ║")
	fmt.Println("╚═════════════════════════════╝")

	// Print hash map entries
	for key, node := range c.CacheMap {
		fmt.Printf("  %s │ Value: %s │ Queue Position: ", key, node.Data)

		// Find the position of the node in the queue
		position := 1
		currentNode := c.Queue.Head
		for currentNode != nil {
			if currentNode == node {
				break
			}
			currentNode = currentNode.Next
			position++
		}

		fmt.Println(position)
	}

	fmt.Println("╔═════════════════════════════╗")
	fmt.Println("║        Queue Order          ║")
	fmt.Println("╚═════════════════════════════╝")

	// Print linked list (queue) entries
	currentNode := c.Queue.Head
	for currentNode != nil {
		fmt.Printf("  %s -> ", currentNode.Data)
		currentNode = currentNode.Next
	}
	fmt.Println("nil\n╚═════════════════════════════╝")
}

// AddToFront adds a new node to the front of the Queue.
func (q *Queue) AddToFront(node *Node) {
	node.Next = q.Head
	if q.Head != nil {
		q.Head.prev = node
	}
	q.Head = node
	if q.tail == nil {
		q.tail = node
	}
}

// MoveToFront moves the given node to the front of the Queue.
func (q *Queue) MoveToFront(node *Node) {
	if node == nil {
		return
	}

	// Remove the node from its current position
	q.RemoveNode(node)

	// Add the node to the front
	q.AddToFront(node)
}

// RemoveNode removes the given node from the Queue.
func (q *Queue) RemoveNode(node *Node) {
	if node == nil {
		return
	}

	// Remove the node from its current position
	if nodePrev := node.prev; nodePrev != nil {
		nodePrev.Next = node.Next
	} else {
		q.Head = node.Next
	}

	if nodeNext := node.Next; nodeNext != nil {
		nodeNext.prev = node.prev
	} else {
		q.tail = node.prev
	}
}

// RemoveFromEnd removes the last node from the Queue.
func (q *Queue) RemoveFromEnd() *Node {
	if q.tail == nil {
		return nil
	}

	removedNode := q.tail
	q.tail = q.tail.prev

	if q.tail != nil {
		q.tail.Next = nil
	} else {
		q.Head = nil
	}

	return removedNode
}

func (c *Cache) Get(key string) ([]byte, error) {
	println(GreenColor+"Getting Cache for Key: ", key, ResetColor)
	node, exists := c.CacheMap[key]
	if !exists {
		fmt.Println(RedColor+"Key: ", key, " not found."+ResetColor)
		c.Misses++ // Increment misses count
		return nil, errors.New("key not found")
	}
	c.Hits++ // Increment hits count

	// Move the accessed node to the front of the Queue
	c.Queue.MoveToFront(node)

	return node.Data.([]byte), nil
}

func (c *Cache) Set(key string, value []byte, duration time.Duration) error {
	println(GreenColor+"Setting cache> Key: ", key, " Value: ", string(value), ResetColor)
	node, exists := c.CacheMap[key]
	c.mutex.Lock()
	if exists {
		// Update existing node
		node.Data = value
		c.Queue.MoveToFront(node)
	} else {
		// Add new node
		newNode := &Node{
			Data: value,
			Next: c.Queue.Head,
			prev: nil,
		}
		c.CacheMap[key] = newNode
		c.Queue.AddToFront(newNode)
		if !c.replayingAOF {
			c.writeToAOF("SET", key, value, duration)
		}
		// Increment cache size
		c.Size++
	}
	c.mutex.Unlock()

	go func() {
		<-time.After(duration)
		err := c.Delete(key)
		if err != nil {
			println("\n"+RedColor+"Error evicting cache with key: ", key, ResetColor)
			return
		}
	}()

	return nil
}

func (c *Cache) Has(key string) bool {
	_, exists := c.CacheMap[key]
	return exists
}

func (c *Cache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	node, exists := c.CacheMap[key]
	if !exists {
		return errors.New("key not found")
	}

	c.Queue.RemoveNode(node)
	delete(c.CacheMap, key)

	c.writeToAOF("DELETE", key, nil, 0)

	// Decrement cache size
	c.Size--

	// Increment evictions count
	c.Evictions++

	return nil
}

// ResetCache resets the cache by clearing the map and Queue.
func (c *Cache) ResetCache() error {
	c.CacheMap = make(map[string]*Node)
	c.Queue = NewQueue()

	c.writeToAOF("FLUSHALL", "", nil, 0)

	// Reset cache statistics
	c.Size = 0
	c.Evictions = 0
	c.Hits = 0
	c.Misses = 0

	return nil
}

func (c *Cache) IsFull() bool {
	// TODO: Implement logic to check if the cache is full
	return false
}

func (c *Cache) IsEmpty() bool {
	return len(c.CacheMap) == 0
}

// GetCacheData returns the current cache data as a map.
func (c *Cache) GetCacheData() map[string][]byte {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cacheData := make(map[string][]byte)
	for key, node := range c.CacheMap {
		cacheData[key] = node.Data.([]byte)
	}
	return cacheData
}

// SetCacheData sets the cache data using the provided map.
func (c *Cache) SetCacheData(cacheData map[string][]byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.CacheMap = make(map[string]*Node)
	c.Queue = NewQueue()

	for key, value := range cacheData {
		newNode := &Node{
			Data: value,
			Next: c.Queue.Head,
			prev: nil,
		}
		c.CacheMap[key] = newNode
		c.Queue.AddToFront(newNode)
	}
}
