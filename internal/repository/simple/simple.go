package simple

import (
	"fmt"
	"sync"
)

type Storage struct {
	m    sync.RWMutex
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

func (s *Storage) Add(id, value string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.data[id] = value
}

func (s *Storage) Get(id string) (string, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	value, ok := s.data[id]
	if !ok {
		return "", fmt.Errorf("не найдено данных для id = %q", id)
	}
	return value, nil
}

func (s *Storage) IDExists(id string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, found := s.data[id]
	return found
}
