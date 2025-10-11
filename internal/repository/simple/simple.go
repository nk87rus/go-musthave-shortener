package simple

import "fmt"

type Storage struct {
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

func (s *Storage) Add(id, value string) {
	s.data[id] = value
}

func (s *Storage) Get(id string) (string, error) {
	value, ok := s.data[id]
	if !ok {
		return "", fmt.Errorf("не найдено данных для id = %q", id)
	}
	return value, nil
}

func (s *Storage) IDExists(id string) bool {
	_, found := s.data[id]
	return found
}
