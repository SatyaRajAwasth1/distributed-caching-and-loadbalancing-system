package cache

import (
	"errors"
	"time"
)

type Cache struct {
	cacheMap map[string]*Node
	queue    *Queue
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

// NewCache creates a new Cache with an empty map and queue.
func New() *Cache {
	return &Cache{
		cacheMap: make(map[string]*Node),
		queue:    NewQueue(),
	}
}

// NewQueue creates a new Queue with empty head and tail.
func NewQueue() *Queue {
	return &Queue{}
}

// AddToFront adds a new node to the front of the queue.
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

// MoveToFront moves the given node to the front of the queue.
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

// RemoveNode removes the given node from the queue.
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

// RemoveFromEnd removes the last node from the queue.
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
	node, exists := c.cacheMap[key]
	if !exists {
		return nil, errors.New("key not found")
	}

	// Move the accessed node to the front of the queue
	c.queue.MoveToFront(node)

	return node.data.([]byte), nil
}

func (c *Cache) Set(key string, value []byte, duration time.Duration) error {
	node, exists := c.cacheMap[key]
	if exists {
		// Update existing node
		node.data = value
		c.queue.MoveToFront(node)
	} else {
		// Add new node
		newNode := &Node{
			data: value,
			next: c.queue.head,
			prev: c.queue.tail,
		}
		c.cacheMap[key] = newNode
		c.queue.AddToFront(newNode)
	}

	go func() {
		<-time.After(duration)
		delete(c.cacheMap, key)
	}()
	return nil
}

func (c *Cache) Has(key string) bool {
	_, exists := c.cacheMap[key]
	return exists
}

func (c *Cache) Delete(key string) error {
	node, exists := c.cacheMap[key]
	if !exists {
		return errors.New("key not found")
	}

	c.queue.RemoveNode(node)
	delete(c.cacheMap, key)

	return nil
}

// ResetCache resets the cache by clearing the map and queue.
func (c *Cache) ResetCache() error {
	c.cacheMap = make(map[string]*Node)
	c.queue = NewQueue()
	return nil
}

func (c *Cache) IsFull() bool {
	// TODO: Implement logic to check if the cache is full
	return false
}

func (c *Cache) IsEmpty() bool {
	return len(c.cacheMap) == 0
}
