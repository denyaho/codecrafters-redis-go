package store

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	
)

type StreamEntry struct {
	ID string
	Value map[string]string
}

type Item struct{
	value any
	expireAt int64
}

type ZSetEntry struct {
	Member string
	Score  float64	
}

type ExpireMap struct {
	data map[string]Item
	mu sync.RWMutex
	signals map[string]chan struct{}
}

func (m *ExpireMap) ZCard(key string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exist := m.data[key]
	sortedSet, ok := item.value.([]ZSetEntry)
	if !exist {
		return 0, nil
	}
	if !ok {
		return 0, ErrWrongType
	}
	return len(sortedSet), nil
}

func (m *ExpireMap) ZScore(key, member string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	item, exist := m.data[key]
	sortedSet, ok := item.value.([]ZSetEntry)
	if !exist {
		return -1, nil
	}
	if !ok {
		return -1, ErrWrongType
	}
	for _, entry := range sortedSet {
		if entry.Member == member {
			return entry.Score, nil
		}
	}
	return -1, nil
}

func (m *ExpireMap) ZRank(key, member string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exist := m.data[key]
	sortedSet, ok := item.value.([]ZSetEntry)
	if !exist {
		return -1, nil
	}
	if !ok {
		return -1, ErrWrongType
	}
	for i, entry := range sortedSet {
		if entry.Member == member {
			return i, nil
		}
	}
	return -1, nil
}

func (m *ExpireMap) ZRange(key string, start, stop int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exist := m.data[key]
	sortedSet, ok := item.value.([]ZSetEntry)
	if !exist {
		return []string{}, nil
	}
	if !ok {
		return []string{}, ErrWrongType
	}

	start, stop = resolveIndex(start, stop, len(sortedSet))
	if start >= len(sortedSet) {
		return []string{}, nil
	}
	if stop >= len(sortedSet) {
		stop = len(sortedSet) - 1
	}
	if start > stop {
		return []string{}, nil
	}
	members := make([]string, stop-start+1)
	for i, entry := range sortedSet[start : stop+1] {
		members[i] = entry.Member
	}
	return members, nil
}

func (m *ExpireMap) ZAdd(key string, score float64, member string) (int, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exist := m.data[key]
	sortedSet, ok := item.value.([]ZSetEntry)

	if exist && !ok {
		return 0, ErrWrongType
	}
	if !exist {
		sortedSet = []ZSetEntry{}
	}
	updated := false
	if exist{
		for i, entry := range sortedSet {
			if entry.Member == member {
				sortedSet[i].Score = score
				updated = true
			}
		}
	}
	if !updated {
		entry := ZSetEntry{
			Member: member,
			Score: score,
		}
		sortedSet = append(sortedSet, entry)
	}
	sort.SliceStable(sortedSet, func(i, j int) bool {
		if sortedSet[i].Score == sortedSet[j].Score {
			return sortedSet[i].Member < sortedSet[j].Member
		}
		return sortedSet[i].Score < sortedSet[j].Score
	})
	m.data[key] = Item{value: sortedSet, expireAt: item.expireAt}
	if updated {
		return 0, nil
	}
	return 1, nil
}


func (m *ExpireMap) Keys(key string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		matched, err := path.Match(key, k)
		if err == nil && matched {
			keys = append(keys, k)
		}
	}
	return keys
}

func (m *ExpireMap) Set(key string, value any, expireAt time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("setting key: %v, value: %v", key, value)

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
	retValue := item.value
	if item.expireAt != 0 && item.expireAt < time.Now().UnixNano() {
		delete(m.data, key)
		return "", false
	}
	return retValue, true
}

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

func (m *ExpireMap) Rpush(key string, values ...string) (int, error) {
	m.mu.Lock()

	item, exist := m.data[key]
	list, ok := item.value.([]string)
	if exist && !ok {
		m.mu.Unlock()
		return 0, ErrWrongType
	} else if !exist {
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
	}else if !exist {
		list = []string{}
	}
	for _, v := range values {
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

func (m *ExpireMap) BLPop(key string, timeout time.Duration) (string, bool, bool) {
	m.mu.Lock()
	list, _ := m.data[key].value.([]string)
	if len(list) > 0 {
		val := list[0]
		m.data[key] = Item{value: list[1:]}
		m.mu.Unlock()
		return val, true, false
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
		return "",false, true
	}
	m.mu.Lock()
	list, _ = m.data[key].value.([]string)
	if len(list) == 0 {
		m.mu.Unlock()
		return "", false, false
	}
	m.data[key] = Item{value: list[1:]}
	m.mu.Unlock()
	return list[0], true, false
}

func (m *ExpireMap) XAdd(key string, entryID string, pairs map[string]string) (error){
	m.mu.Lock()
	item, exist := m.data[key]
	stream, ok := item.value.([]StreamEntry)
	if exist && !ok {
		m.mu.Unlock()
		return ErrWrongType
	}else if !exist {
		stream = []StreamEntry{}
	}
	newEntry := StreamEntry{
		ID: entryID,
		Value: pairs,
	}
	stream = append(stream, newEntry)
	m.data[key] = Item{value: stream}

	ch := m.signals[key]
	m.mu.Unlock()
	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	return nil
}

func _isGreaterStreamID(id1, id2 string) (bool) {
	id1_split:= strings.Split(id1, "-")
	id2_split := strings.Split(id2, "-")
	if id1_split[0] > id2_split[0] {
		return true
	} else if id1_split[0] == id2_split[0] && id1_split[1] > id2_split[1] {
		return true
	}
	return false
}

func (m *ExpireMap) XReadBlock(key, entryID string, timeout time.Duration) (bool, bool) {
	m.mu.Lock()
	stream, _ := m.data[key].value.([]StreamEntry)
	if len(stream) > 0  && entryID != "$" {
		lastStream := stream[len(stream)-1]
		if _isGreaterStreamID(entryID, lastStream.ID) {
			m.mu.Unlock()
			return true,false
		}
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
			return false, true
	}
	m.mu.Lock()
	stream, _ = m.data[key].value.([]StreamEntry)
	if len(stream) == 0 {
		m.mu.Unlock()
		return false, false
	}
	m.mu.Unlock()
	return true, false
}

type Store interface {
	Set(key, value string, expireAt time.Duration)
	Get(key string) (string, bool)
	Rpush(key string, value ...string) int
	Lrange(key string, start, stop int) []string
	LPush(key string, value ...string) int
	Llen(key string) int
	Lpop(key string) (string, bool)
	BLPop(key string, timeout time.Duration) (string, bool, bool)
	XAdd(key string, entryID string, pairs map[string]string) error
	XReadBlock(key string, timeout time.Duration) ([]StreamEntry, bool, bool)
}

func NewExpireMap() *ExpireMap {
	return &ExpireMap{data: make(map[string]Item), signals: make(map[string]chan struct{})}
}