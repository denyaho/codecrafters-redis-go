package store

import "sort"

type GeoEntry struct {
	Longitude float64
	Latitude  float64
	Member    string
}

func (m *ExpireMap) GeoAdd(key string, longitude, latitude float64, member string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exist := m.data[key]
	geoSet, ok := item.value.([]GeoEntry)
	if exist && !ok {
		return 0, ErrWrongType
	}
	if !exist {
		geoSet = []GeoEntry{}
	}
	updated := false

	if exist {
		for i, entry := range geoSet {
			if entry.Member == member {
				geoSet[i].Longitude = longitude
				geoSet[i].Latitude = latitude
				updated = true
				break
			}
		}
	}
	if !updated {
		entry := GeoEntry{
			Longitude: longitude,
			Latitude:  latitude,
			Member:    member,
		}
		geoSet = append(geoSet, entry)
	}
	sort.SliceStable(geoSet, func(i, j int) bool {
		if geoSet[i].Member == geoSet[j].Member {
			return geoSet[i].Longitude < geoSet[j].Longitude
		}
		return geoSet[i].Member < geoSet[j].Member
	})
	m.data[key] = Item{value: geoSet, expireAt: item.expireAt}
	if updated {
		return 0, nil
	}
	return 1, nil
}