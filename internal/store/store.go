package store

import (
	"sync"
	"time"
)

type Item struct{
	value string
	expireAt int64
}

type ExpireMap struct {
	data map[string]Item
	mu sync.RWMutex
}

func (m *ExpireMap) Set(key, value string, expireAt time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var exp int64
	if expireAt == 0 {
		exp = 0
	}else {
		exp = time.Now().Add(expireAt).UnixNano()
	}

	m.data[key] = Item{
		value: value,
		expireAt: exp,
	}
}

func (m *ExpireMap) Get(key string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.data[key]
	if !ok {
		return "", false
	}
	if item.expireAt != 0 && item.expireAt < time.Now().UnixNano() {
		delete(m.data, key)
		return "", false
	}
	return item.value, true
}

type Store interface {
	Set(key, value string, expireAt time.Duration)
	Get(key string) (string, bool)
}

func NewExpireMap() *ExpireMap {
	return &ExpireMap{data: make(map[string]Item)}
}