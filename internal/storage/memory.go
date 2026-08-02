package storage

import (
	"errors"
)

type MemoryStorage struct {
	data map[string][]byte
}

func NewMemeoryStore() *MemoryStorage {
	return &MemoryStorage{ data: make(map[string][]byte) }
}

func (m *MemoryStorage) Set(key string, value []byte) error {
	m.data[key] = value
	
	return nil
}

func (m *MemoryStorage)	Get(key string) ([]byte, error) {
	value, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	copyValue := make([]byte, len(value))
	copy(copyValue, value)
	return copyValue, nil
}

func (m *MemoryStorage)	Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MemoryStorage)	Exists(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MemoryStorage) Keys() []string {
	keys := make([]string, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}