package store

import (
	"sync"
	"time"
)

type Store interface {
	Add(key, val string, ttl int64)
	Set(key, val string, ttl int64)
	Get(key string) (string, bool)
	Keys(key string) []string
	FlushAll()
	SnapShot() SnapShot
}

type memory struct {
	mu     sync.RWMutex
	data   map[string]string
	expiry map[string]int64
}

func New() Store {
	return &memory{
		data:   make(map[string]string),
		expiry: make(map[string]int64),
	}
}

func (m *memory) Add(key string, val string, ttl int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ttl == 0 {
		m.data[key] = val
		delete(m.expiry, key)
	} else if ttl > time.Now().UnixMilli() {
		m.data[key] = val
		m.expiry[key] = ttl
	}
}

func (m *memory) Set(key string, val string, ttl int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ttl < 0 {
		return
	}

	m.data[key] = val

	if ttl > 0 {
		m.expiry[key] = time.Now().UnixMilli() + ttl
	} else if ttl == 0 {
		delete(m.expiry, key)
	}
}

func (m *memory) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if exp, ok := m.expiry[key]; ok && time.Now().UnixMilli() > exp {
		delete(m.data, key)
		delete(m.expiry, key)
		return "", false
	}

	val, ok := m.data[key]
	return val, ok
}

func (m *memory) Keys(key string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0)

	for k, _ := range m.data {
		keys = append(keys, k)
	}

	return keys
}

func (m *memory) FlushAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]string)
	m.expiry = make(map[string]int64)
}
