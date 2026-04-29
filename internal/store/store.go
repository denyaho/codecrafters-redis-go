package store

import (
	"sync"
	"time"
	"errors"
	"fmt"
)

type Item struct{
	value any
	expireAt int64
}

type ExpireMap struct {
	data map[string]Item
	mu sync.RWMutex
	signals map[string]chan struct{}
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

func (m *ExpireMap) Get(key string) (any, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.data[key]
	if !ok {
		return "", false
	}
	str, ok := item.value.(string)
	if !ok {
		return "", false
	}

	if item.expireAt != 0 && item.expireAt < time.Now().UnixNano() {
		delete(m.data, key)
		return "", false
	}
	return str, true
}

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

func (m *ExpireMap) Rpush(key string, values ...string) (int, error) {
	m.mu.Lock()

	item, exist := m.data[key]
	list, ok := item.value.([]string)
	if exist && !ok {
		m.mu.Unlock()
		return 0, ErrWrongType
	} else{
		list = []string{}
	}
	for _, v := range values {
		list = append(list, v)
	}
	m.data[key] = Item{value: list, expireAt: item.expireAt}
	length := len(list)

	ch := m.signals[key]
	m.mu.Unlock()
	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}	
	}
	return length, nil
}

func (m *ExpireMap) LPush(key string, values ...string) (int, error) {
	m.mu.Lock()
	item, exist := m.data[key]
	list, ok := item.value.([]string)
	if exist && !ok {
		m.mu.Unlock()
		return 0, ErrWrongType
	} else{
		list = []string{}
	}
	for _, v := range values {
		fmt.Printf("pushing %s to list\n", v)
		list = append([]string{v}, list...)
	}
	m.data[key] = Item{value: list, expireAt: item.expireAt}
	length := len(list)

	ch := m.signals[key]
	m.mu.Unlock()
	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	fmt.Printf("%v", list)
	return length, nil
}

func resolveIndex(start, stop, length int) (int, int) {
	if start < 0 {
		start = length + start
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

func (m *ExpireMap) Lrange(key string, start, stop int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, exist := m.data[key]
	if !exist {
		return []string{}, nil
	}
	list, ok := item.value.([]string)
	if !ok {
		return []string{}, ErrWrongType
	}
	start, stop = resolveIndex(start, stop, len(list))
	if start >= len(list) {
		return []string{}, nil
	}
	if stop >= len(list) {
		stop = len(list) - 1
	}
	if start > stop {
		return []string{}, nil
	}
	return list[start:stop+1], nil
}

func (m *ExpireMap) Llen(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, exist := m.data[key]
	if !exist {
		return 0, nil
	}
	list, ok := item.value.([]string)
	if !ok {
		return 0, ErrWrongType
	}
	if list == nil {
		return 0, nil
	}
	return len(list), nil
}



func (m *ExpireMap) Lpop(key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, exist := m.data[key]
	if !exist {
		return "", false, nil // key does not exist
	}
	list, ok := item.value.([]string)
	if !ok {
		return "", false, ErrWrongType
	}
	if len(list) == 0 {
		return "", false, nil // list is empty
	}
	m.data[key] = Item{value: list[1:], expireAt: item.expireAt}
	return list[0], true, nil
}

func (m *ExpireMap) BLPop(key string, timeout time.Duration) (string, bool) {
	m.mu.Lock()
	list, _ := m.data[key].value.([]string)
	if len(list) > 0 {
		val := list[0]
		m.data[key] = Item{value: list[1:]}
		m.mu.Unlock()
		fmt.Printf("popped %s from list\n", val)
		return val, true
	}

	ch := make(chan struct{}, 1)
	m.signals[key] = ch
	m.mu.Unlock()

	var timer <- chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	}
	select {
	case <-ch:
	case <-timer:
		return "",false
	}
	m.mu.Lock()
	list, _ = m.data[key].value.([]string)
	if len(list) == 0 {
		m.mu.Unlock()
		return "", false
	}
	m.data[key] = Item{value: list[1:]}
	fmt.Printf("popped %s from list\n", list[0])
	m.mu.Unlock()
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
	BLPop(key string, timeout time.Duration) (string, bool)
}

func NewExpireMap() *ExpireMap {
	return &ExpireMap{data: make(map[string]Item), signals: make(map[string]chan struct{})}
}