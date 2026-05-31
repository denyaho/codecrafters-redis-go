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

func (m *Manager) Unsubscribe(client *Client, channel string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.channels[channel]
	var index int
	for i, c := range clients {
		if c.ID == client.ID {
			index = i
			break
		}
	}
	m.channels[channel] = append(clients[:index], clients[index+1:]...)
}

func (m *Manager) Publish(channel,message string) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, client := range m.channels[channel]{
		client.Connection.Write(resp.BuildArrayForPUBLISH(channel, message))
		count++
	}
	return resp.BuildInteger(count)
}