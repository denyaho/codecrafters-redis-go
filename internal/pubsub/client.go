package pubsub

import (
	"net"
	"sync/atomic"
)
//struct -> メモリを消費しない型

type Client struct {
	ID                 int64
	Connection         net.Conn
	SubscribedChannels map[string]struct{}
	SubscriptionCount  int
	IsSubscribed bool
}

var id int64

func NewClient(conn net.Conn) *Client {
	return &Client{
		ID:                 atomic.AddInt64(&id, 1),
		Connection:         conn,
		SubscribedChannels: make(map[string]struct{}),
		SubscriptionCount:  0,
		IsSubscribed: false,
	}
}

func (c *Client) Subscribe(channel string) {
	if _, exists := c.SubscribedChannels[channel]; !exists {
		c.SubscribedChannels[channel] = struct{}{}
		c.SubscriptionCount++
	}
}

func (c *Client) Unsubscribe(channel string) {
	if _, exists := c.SubscribedChannels[channel]; exists {
		delete(c.SubscribedChannels, channel)
		c.SubscriptionCount--
		if c.SubscriptionCount == 0 {
			c.IsSubscribed = false
		}
	}
}
