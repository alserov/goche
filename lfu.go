package goche

import (
	"context"
	"sync"
	"time"
)

func NewLFU(mods ...ModifierFunc) Cache {
	ctx, cancel := context.WithCancel(context.Background())

	c := lfu{
		limit:         DefaultLimit,
		vals:          make(map[string]any),
		pushCh:        make(chan *lfuNode, DefaultLimit),
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
	go c.clearAfterPeriod()

	return &c
}

type lfuNode struct {
	val any
	key string

	new bool

	freq uint64

	lastUsedAt time.Time

	prev *lfuNode
	next *lfuNode
}

type lfu struct {
	head *lfuNode
	tail *lfuNode

	mu sync.RWMutex

	vals map[string]any

	stopCtx       context.Context
	stopCtxCancel context.CancelFunc

	popCh       chan struct{}
	pushCh      chan *lfuNode
	updatePosCh chan string

	len   int
	limit int

	clearPeriod *time.Duration
}

func (c *lfu) Close() {
	c.stopCtxCancel()
}

func (c *lfu) Get(ctx context.Context, key string) (any, bool) {
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

func (c *lfu) Set(ctx context.Context, key string, val any) {
	c.mu.Lock()
	_, ok := c.vals[key]
	if ok {
		c.mu.Unlock()
		return
	}
	c.vals[key] = val
	c.mu.Unlock()

	select {
	case c.pushCh <- newLFUNode(key, val):
	case <-c.stopCtx.Done():
		close(c.pushCh)
	}
}

func (c *lfu) updatePos() {
	upsert := func(key string) {
		c.mu.Lock()
		defer c.mu.Unlock()

		var n *lfuNode
		for n = c.head; n != nil && n.key != key; n = n.next {
		}

		if n != nil {
			n.lastUsedAt = time.Now()
			n.freq++

			currentNode := n.prev
			for currentNode != nil && currentNode.freq <= n.freq {
				currentNode = currentNode.prev
			}

			if currentNode == nil {
				for n.prev != nil && n.prev.freq <= n.freq {
					c.swap(n.prev, n)
				}
			} else {
				if currentNode.next == n {
					c.swap(currentNode, n)
				} else {
					if currentNode.next != nil {
						currentNode.next.prev = n
					}
					n.next = currentNode.next
					n.prev = currentNode
					currentNode.next = n
				}
			}
		}
	}

	for key := range c.updatePosCh {
		upsert(key)
	}
}

func (c *lfu) push() {
	defer func() {
		close(c.popCh)
	}()

	insert := func(n *lfuNode) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.head == nil {
			c.head = n
			c.tail = n
		} else {
			if c.len >= c.limit {
				c.popCh <- struct{}{}
			}

			currentNode := c.tail
			for currentNode != nil && currentNode.freq <= n.freq {
				currentNode = currentNode.prev
			}

			if currentNode == nil {
				c.head.prev = n
				n.next = c.head
				c.head = n
			} else {
				if currentNode.next != nil {
					currentNode.next.prev = n
				}
				n.next = currentNode.next
				n.prev = currentNode
				currentNode.next = n
			}
		}

		c.len++
	}

	for node := range c.pushCh {
		insert(node)
	}
}

func (c *lfu) pop() {
	extrude := func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.vals, c.tail.key)

		if c.tail != nil {
			c.tail = c.tail.prev
			if c.tail != nil {
				c.tail.next = nil
			}
		}
	}

	for range c.popCh {
		extrude()
	}
}

func (c *lfu) swap(prev, curr *lfuNode) bool {
	if prev == nil || curr == nil {
		return false
	}

	prev.next = curr.next
	if curr.next != nil {
		curr.next.prev = prev
	} else {
		c.tail = prev
	}

	curr.prev = prev.prev
	if prev.prev != nil {
		prev.prev.next = curr
	} else {
		c.head = curr
	}

	curr.next = prev
	prev.prev = curr

	return true
}

func (c *lfu) clearAfterPeriod() {
	if c.clearPeriod != nil {
		for range time.Tick(*c.clearPeriod) {
			c.mu.Lock()
			if c.head != nil {
				curr := c.head
				for curr != nil {
					if time.Now().UnixMilli()-curr.lastUsedAt.UnixMilli() > c.clearPeriod.Milliseconds() {
						delete(c.vals, curr.key)

						if curr.prev == nil {
							c.head = c.head.next
						} else {
							curr.prev.next = curr.next
						}

						if curr.next == nil {
							c.tail = c.tail.prev
						} else {
							curr.next.prev = curr.prev
						}
					}
					curr = curr.next
				}
			}
			c.mu.Unlock()
		}
	}
}

func newLFUNode(key string, val any) *lfuNode {
	return &lfuNode{val: val, key: key, lastUsedAt: time.Now()}
}
