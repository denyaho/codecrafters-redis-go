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

func (m *ExpireMap) LPush(key string, value ...string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range value {
		m.lists[key] = append([]string{v}, m.lists[key]...)
	}
	return len(m.lists[key])
}

func resolveIndex(start, stop, length int) (int, int) {
	if start < 0 {
		start = length * start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = length + stop
		if stop < 0 {
			stop = 0
		}
	}
	return start, stop
}

func (m *ExpireMap) Lrange(key string, start, stop int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list, ok := m.lists[key]
	if !ok {
		return []string{}
	}
	start, stop = resolveIndex(start, stop, len(list))
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

func (m *ExpireMap) Llen(key string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list, ok := m.lists[key]
	if !ok {
		return 0
	}
	return len(list)
}

func (m *ExpireMap) Lpop(key string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list, ok := m.lists[key]
	if !ok || len(list) == 0{
		return "", false
	}
	m.lists[key] = m.lists[key][1:]
	return list[0], true
}

type Store interface {
	Set(key, value string, expireAt time.Duration)
	Get(key string) (string, bool)
	Rpush(key string, value ...string) int
	Lrange(key string, start, stop int) []string
	LPush(key string, value ...string) int
	Llen(key string) int
	Lpop(key string) (string, bool)
}

func NewExpireMap() *ExpireMap {
	return &ExpireMap{data: make(map[string]Item), lists: make(map[string][]string)}
}