package store

import "time"

type SnapShot struct {
	Data   map[string]string
	Expiry map[string]int64
}

func (m *memory) SnapShot() SnapShot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := make(map[string]string, len(m.data))
	expiry := make(map[string]int64, len(m.expiry))

	now := time.Now().UnixMilli()

	for k, v := range m.data {
		if ttl, ok := m.expiry[k]; ok {
			if ttl <= now {
				delete(m.data, k)
				delete(m.expiry, k)
				continue
			}
			expiry[k] = ttl
		}
		data[k] = v
	}

	return SnapShot{
		Data:   data,
		Expiry: expiry,
	}
}
