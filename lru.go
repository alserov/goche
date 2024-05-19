package goche

import (
	"context"
	"sync"
)

func NewLRU(mods ...ModifierFunc) Cache {
	ctx, cancel := context.WithCancel(context.Background())

	c := lru{
		limit:         DefaultLimit,
		vals:          make(map[string]any),
		pushCh:        make(chan *lruNode, DefaultLimit),
		popCh:         make(chan struct{}, 10),
		updatePosCh:   make(chan string, 10),
		stopCtx:       ctx,
		stopCtxCancel: cancel,
	}

	for _, mod := range mods {
		mod(&c)
	}

	go c.pop()
	go c.push()
	go c.updatePos()

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

	vals map[string]any

	mu sync.RWMutex

	stopCtx       context.Context
	stopCtxCancel context.CancelFunc

	popCh       chan struct{}
	pushCh      chan *lruNode
	updatePosCh chan string

	len   uint64
	limit uint64
}

func (c *lru) Close() {
	c.stopCtxCancel()
}

// Get retrieves value by its id, returns false if not found
func (c *lru) Get(ctx context.Context, key string) (any, bool) {
	c.mu.RLock()
	val, ok := c.vals[key]
	c.mu.RUnlock()

	if ok {
		select {
		case c.updatePosCh <- key:
		case <-ctx.Done():
			close(c.updatePosCh)
		}
	}

	return val, ok
}

func (c *lru) Set(ctx context.Context, key string, val any) {
	c.mu.Lock()
	_, ok := c.vals[key]
	if ok {
		c.mu.Unlock()
		return
	}
	c.vals[key] = val
	c.mu.Unlock()

	select {
	case c.pushCh <- newLRUNode(key, val):
	case <-c.stopCtx.Done():
		close(c.pushCh)
	}
}

func (c *lru) pop() {
	deleteAtEnd := func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.vals, c.tail.key)
		c.tail = c.tail.prev
		c.tail.next = nil
	}

	for range c.popCh {
		deleteAtEnd()
	}
}

func (c *lru) push() {
	defer func() {
		close(c.popCh)
	}()

	insertAtStart := func(n *lruNode) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.head == nil {
			c.head = n
			c.tail = n
		} else {
			if c.len >= c.limit {
				c.popCh <- struct{}{}
			}
			n.next = c.head
			c.head.prev = n
			c.head = n
		}

		c.len++
	}

	for n := range c.pushCh {
		insertAtStart(n)
	}
}

func (c *lru) updatePos() {
	removeAndInsertAtStart := func(key string) {
		c.mu.Lock()
		defer c.mu.Unlock()

		curr := c.head

		for curr != nil {
			if curr.key == key && curr.prev != nil {
				if curr.next != nil {
					curr.next.prev = curr.prev
				} else {
					c.tail = curr.prev
				}

				curr.prev.next = curr.next

				curr.prev = nil
				curr.next = c.head
				c.head.prev = curr
				c.head = curr

				break
			}

			curr = curr.next
		}
	}

	for key := range c.updatePosCh {
		removeAndInsertAtStart(key)
	}
}

func newLRUNode(key string, val any) *lruNode {
	return &lruNode{val: val, key: key}
}
