package pubsub

import (
	"sync"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Manager struct {
	mu sync.Mutex
	channels map[string][]*Client
}

func NewManager() *Manager {
	return &Manager {
		channels: make(map[string][]*Client),
	}
}

func (m *Manager) Subscribe(client *Client, channel string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channels[channel] = append(m.channels[channel], client)
}

func (m *Manager) Publish(channel,message string) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, client := range m.channels[channel]{
		client.Connection.Write([]byte(message))
		count++
	}
	return resp.BuildInteger(count)
}