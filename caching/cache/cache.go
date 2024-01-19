package cache

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Cache struct {
	CacheMap       map[string]*Node
	Queue          *Queue
	mutex          sync.Mutex
	aofFile        *os.File
	aofWriteTicker *time.Ticker
	aofMutex       sync.Mutex
}

type Queue struct {
	head *Node
	tail *Node
}

type Node struct {
	data interface{}
	prev *Node
	next *Node
}

// NewCache creates a new Cache with an empty map and Queue.
func NewCache(aofFilePath string) *Cache {
	cache := &Cache{
		CacheMap: make(map[string]*Node),
		Queue:    NewQueue(),
		mutex:    sync.Mutex{},
	}

	// Open or create AOF file
	aofFile, err := openOrCreateAOFFile(aofFilePath)
	if err != nil {
		fmt.Println(RedColor+"Error opening AOF file:", err, ResetColor)
		return nil
	}

	cache.aofFile = aofFile

	// Replay AOF file
	cache.ReplayAOF(aofFilePath)

	return cache
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
		fmt.Printf("  %s │ Value: %s │ Queue Position: ", key, node.data)

		// Find the position of the node in the queue
		position := 1
		currentNode := c.Queue.head
		for currentNode != nil {
			if currentNode == node {
				break
			}
			currentNode = currentNode.next
			position++
		}

		fmt.Println(position)
	}

	fmt.Println("╔═════════════════════════════╗")
	fmt.Println("║        Queue Order          ║")
	fmt.Println("╚═════════════════════════════╝")

	// Print linked list (queue) entries
	currentNode := c.Queue.head
	for currentNode != nil {
		fmt.Printf("  %s -> ", currentNode.data)
		currentNode = currentNode.next
	}
	fmt.Println("nil\n╚═════════════════════════════╝")
}

// AddToFront adds a new node to the front of the Queue.
func (q *Queue) AddToFront(node *Node) {
	node.next = q.head
	if q.head != nil {
		q.head.prev = node
	}
	q.head = node
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
	if nodePrev := node.prev; nodePrev != nil {
		nodePrev.next = node.next
	} else {
		q.head = node.next
	}

	if nodeNext := node.next; nodeNext != nil {
		nodeNext.prev = node.prev
	} else {
		q.tail = node.prev
	}

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
		nodePrev.next = node.next
	} else {
		q.head = node.next
	}

	if nodeNext := node.next; nodeNext != nil {
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
		q.tail.next = nil
	} else {
		q.head = nil
	}

	return removedNode
}

func (c *Cache) Get(key string) ([]byte, error) {
	println(GreenColor+"Getting Cache for Key: ", key, ResetColor)
	node, exists := c.CacheMap[key]
	if !exists {
		fmt.Println(RedColor+"Key: ", key, " not found."+ResetColor)
		return nil, errors.New("key not found")
	}

	// Move the accessed node to the front of the Queue
	c.Queue.MoveToFront(node)

	return node.data.([]byte), nil
}

func (c *Cache) Set(key string, value []byte, duration time.Duration) error {
	println(GreenColor+"Setting cache> Key: ", key, " Value: ", string(value), ResetColor)
	node, exists := c.CacheMap[key]
	if exists {
		// Update existing node
		c.mutex.Lock()
		node.data = value
		c.Queue.MoveToFront(node)
		c.mutex.Unlock()
	} else {
		// Add new node
		newNode := &Node{
			data: value,
			next: c.Queue.head,
			prev: nil,
		}
		c.mutex.Lock()
		c.CacheMap[key] = newNode
		c.Queue.AddToFront(newNode)
		c.mutex.Unlock()
	}

	go func() {
		<-time.After(duration)
		err := c.Delete(key)
		if err != nil {
			println(RedColor+"Error evicting cache with key: ", key, GreenColor)
			return
		}
	}()

	// Write to AOF file
	c.writeToAOF("SET", key, value)

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

	// Write to AOF file
	c.writeToAOF("DELETE", key, nil)

	return nil
}

// ResetCache resets the cache by clearing the map and Queue.
func (c *Cache) ResetCache() error {
	c.CacheMap = make(map[string]*Node)
	c.Queue = NewQueue()
	return nil
}

func (c *Cache) IsFull() bool {
	// TODO: Implement logic to check if the cache is full
	return false
}

func (c *Cache) IsEmpty() bool {
	return len(c.CacheMap) == 0
}
