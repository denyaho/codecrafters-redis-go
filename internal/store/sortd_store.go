package store

import (
	"sort"
)

type ZSetEntry struct {
	Member string
	Score  float64
}

func (m *ExpireMap) ZGet(key, member string) (float64, error) {
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

func (m *ExpireMap) ZRem(key, member string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exist := m.data[key]
	sortedSet, ok := item.value.([]ZSetEntry)
	if !exist {
		return 0, nil
	}
	if !ok {
		return 0, ErrWrongType
	}

	count := 0
	for i, value := range sortedSet {
		if value.Member == member {
			sortedSet = append(sortedSet[:i], sortedSet[i+1:]...)
			count ++
		}
	}
	m.data[key] = Item{value: sortedSet, expireAt: item.expireAt}
	return count, nil
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
	if exist {
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
			Score:  score,
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
