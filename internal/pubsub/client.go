package pubsub

import (
	"net"
	"sync/atomic"
	"sync"
)
//struct -> メモリを消費しない型

type Client struct {
	ID                 int64
	Connection         net.Conn
	SubscribedChannels map[string]struct{}
	SubscribedPatterns map[string]struct{}
	SubscriptionCount  int
	IsSubscribed bool
	mu sync.RWMutex
}

var id int64

func NewClient(conn net.Conn) *Client {
	return &Client{
		ID:                 atomic.AddInt64(&id, 1),
		Connection:         conn,
		SubscribedChannels: make(map[string]struct{}),
		SubscribedPatterns: make(map[string]struct{}),
		SubscriptionCount:  0,
		IsSubscribed: false,
	}
}

func (c *Client) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.SubscribedChannels[channel]; !exists {
		c.SubscribedChannels[channel] = struct{}{}
		c.SubscriptionCount++
		c.IsSubscribed = true
	}
}

func (c *Client) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.SubscribedChannels[channel]; exists {
		delete(c.SubscribedChannels, channel)
		c.SubscriptionCount--
		if c.SubscriptionCount == 0 {
			c.IsSubscribed = false
		}
	}
}

func (c *Client) PSubscribe(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.SubscribedPatterns[pattern]; !exists {
		c.SubscribedPatterns[pattern] = struct{}{}
		c.SubscriptionCount++
		c.IsSubscribed = true
	}
}

func (c *Client) PUnsubscribe(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.SubscribedPatterns[pattern]; exists {
		delete(c.SubscribedPatterns, pattern)
		c.SubscriptionCount--
		if c.SubscriptionCount == 0 {
			c.IsSubscribed = false
		}
	}
}
