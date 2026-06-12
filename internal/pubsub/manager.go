package pubsub

import (
	"sync"
	"path"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Manager struct {
	mu sync.Mutex
	channels map[string][]*Client
	patterns map[string][]*Client
	Users map[string]*UserInfo
}

func NewManager() *Manager {
	m := &Manager {
		channels: make(map[string][]*Client),
		patterns: make(map[string][]*Client),
		Users: make(map[string]*UserInfo),
	}
	m.Users["default"] = &UserInfo{
		Username: "default",
		Passwords: []string{},
		Flags: []string{"nopass"},
		Selectors: Selectors{
			Commands: []string{"+@all"},
			Keys: []string{"~*"},
			Channels: []string{"&*"},
		},
	}
	return m
}

func (m *Manager) GetUser(username string) *UserInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user, exists := m.Users[username]; exists {
		return user
	}
	return nil
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
	for i, c := range clients {
		if c.ID == client.ID {
			m.channels[channel] = append(clients[:i], clients[i+1:]...)
			if len(m.channels[channel]) == 0 {
				delete(m.channels, channel)
			}
			return 
		}
	}
}

func (m *Manager) PSubscribe(client *Client, pattern string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.patterns[pattern] = append(m.patterns[pattern], client)
}

func (m *Manager) PUnsubscribe(client *Client, pattern string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.patterns[pattern]
	for i, c := range clients {
		if c.ID == client.ID {
			m.patterns[pattern] = append(clients[:i], clients[i+1:]...)
			if len(m.patterns[pattern]) == 0 {
				delete(m.patterns, pattern)
			}
			return 
		}
	}
}


func (m *Manager) Publish(channel,message string) []byte {
	m.mu.Lock()
	targes := make([]*Client, len(m.channels[channel]))
	copy(targes, m.channels[channel])

	type patternClient struct {
		client *Client
		pattern string
	}
	patternTargets := []patternClient{}
	for pattern, clients := range m.patterns {
		if match, _ := path.Match(pattern, channel); match {
			for _, client := range clients{
				patternTargets = append(patternTargets, patternClient{client: client, pattern: pattern})
			}
		}
	}

	m.mu.Unlock()

	count := 0
	for _, client := range targes{
		client.Connection.Write(resp.BuildArrayForPUBLISH(channel, message))
		count++
	}
	for _, pn := range patternTargets {
		pn.client.Connection.Write(resp.BuildArrayForPUBLISH(pn.pattern, message))
		count++
	}

	return resp.BuildInteger(count)
}