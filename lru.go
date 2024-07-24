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
	if !ok {
		return nil, false
	}

	if len(c.vals) == 2 {
		c.head, c.tail = c.tail, c.head
		c.head.next = c.tail
		c.tail.prev = c.head
	} else {
		if val.prev != nil {
			c.vals[key].prev.next = c.vals[key].next
		}

		if val.next != nil {
			c.vals[key].next.prev = c.vals[key].prev
		} else {
			if c.tail != nil {
				c.tail = c.tail.prev
			}
		}
	}
	c.mu.RUnlock()

	return val.val, true
}

func (c *lru) Set(ctx context.Context, key string, val any) {
	c.mu.Lock()
	node := &lruNode{
		prev: nil,
		next: c.head,
		key:  key,
		val:  val,
	}

	if len(c.vals) == 0 {
		c.tail = node
	}

	c.vals[key] = node
	c.head = c.vals[key]

	if len(c.vals) == c.limit {
		if c.tail != nil {
			delete(c.vals, c.tail.key)
			c.tail = c.tail.prev
		}
	}
	c.mu.Unlock()
}
