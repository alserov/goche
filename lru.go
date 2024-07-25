package goche

import (
	"context"
	"sync"
)

func NewLRU(mods ...ModifierFunc) Cache {
	c := lru{
		limit: DefaultLimit,
		vals:  make(map[string]*lruNode),
	}

	for _, mod := range mods {
		mod(&c)
	}

	return &c
}

type lruNode struct {
	prev *lruNode
	next *lruNode

	key string
	val any
}

type lru struct {
	head *lruNode
	tail *lruNode

	vals map[string]*lruNode

	mu sync.RWMutex

	limit int
}

func (c *lru) Close() {
	clear(c.vals)
	c.head = nil
	c.tail = nil
}

// Get retrieves value by its id, returns false if not found
func (c *lru) Get(ctx context.Context, key string) (any, bool) {
	c.mu.RLock()
	val, ok := c.vals[key]
	c.mu.RUnlock()

	if ok {
		c.updatePos(val)
		return val.val, true
	}

	return nil, false
}

func (c *lru) updatePos(val *lruNode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if val != c.head {
		if len(c.vals) == 2 {
			c.head, c.tail = c.tail, c.head
			c.head.next = c.tail
			c.tail.prev = c.head
		} else {
			if val.prev != nil {
				val.prev.next = val.next
			}
			if val.next != nil {
				val.next.prev = val.prev
			}

			if val == c.tail {
				c.tail = c.tail.prev
			}

			val.next = c.head
			c.head.prev = val
			c.head = val
		}
	}
}

func (c *lru) Set(ctx context.Context, key string, val any) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.vals[key]; ok {
		return
	}

	node := &lruNode{
		key: key,
		val: val,
	}

	c.vals[key] = node

	if c.tail == nil {
		c.tail = node
	}
	if c.head != nil {
		c.head.prev = node
		node.next = c.head
	}

	c.head = node

	if len(c.vals) == c.limit {
		if c.tail != nil {
			delete(c.vals, c.tail.key)
			c.tail = c.tail.prev
		} else {
			c.head = nil
		}
	}
}
