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
	lists map[string][]string
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

func (m *ExpireMap) Rpush(key string, value ...string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.lists[key]
	list = append(list, value...)
	m.lists[key] = list
	return len(list)
}

func (m *ExpireMap) Lrange(key string, start, stop int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list, ok := m.lists[key]
	if !ok {
		return []string{}
	}
	if start >= len(list) {
		return []string{}
	}
	if stop >= len(list) {
		stop = len(list) - 1
	}
	if start > stop {
		return []string{}
	}
	return list[start:stop+1]
}

type Store interface {
	Set(key, value string, expireAt time.Duration)
	Get(key string) (string, bool)
	Rpush(key string, value ...string) int
	Lrange(key string, start, stop int) []string
}

func NewExpireMap() *ExpireMap {
	return &ExpireMap{data: make(map[string]Item), lists: make(map[string][]string)}
}