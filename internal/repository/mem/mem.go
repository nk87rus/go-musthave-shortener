package memstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"strconv"
	"sync"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/rs/zerolog/log"
)

//go:generate go run github.com/vektra/mockery/v2 --name=ExtStorage --inpackage --testonly
type ExtStorage interface {
	LoadData(any) error
	SaveData(iter.Seq[model.StorageRecord]) error
}

type MemStorage struct {
	m          sync.RWMutex
	data       map[string]model.StorageRecord
	lastUUID   int
	extStorage ExtStorage
}

func NewStorage(extStorage ExtStorage) (*MemStorage, error) {
	// newFS, err := filestorage.NewFileStorage(filePath)
	// if err != nil {
	// 	return nil, err
	// }

	var newStorage = MemStorage{
		lastUUID:   0,
		data:       make(map[string]model.StorageRecord),
		extStorage: extStorage,
	}

	if err := newStorage.extStorage.LoadData(&newStorage); err != nil {
		return nil, err
	}

	return &newStorage, nil
}

func (s *MemStorage) UnmarshalJSON(data []byte) error {
	var tmpData []model.StorageRecord
	if err := json.Unmarshal(data, &tmpData); err != nil {
		return err
	}

	s.data = make(map[string]model.StorageRecord, len(tmpData))
	for _, rec := range tmpData {
		uuid, err := strconv.Atoi(rec.UUID)
		if err != nil {
			log.Err(err).Any("record", rec).Msg("не корректный формат uuid")
			continue
		}
		s.data[rec.ShortURL] = rec
		if s.lastUUID < uuid {
			s.lastUUID = uuid
		}
	}

	return nil
}

func (s *MemStorage) Add(ctx context.Context, sURL, oURL string) error {
	s.m.Lock()
	defer s.m.Unlock()
	s.lastUUID++
	s.data[sURL] = model.StorageRecord{UUID: strconv.Itoa(s.lastUUID), ShortURL: sURL, OrigURL: oURL}
	// save to file
	if err := s.extStorage.SaveData(maps.Values(s.data)); err != nil {
		return err
	}
	return nil
}

func (s *MemStorage) Get(ctx context.Context, sURL string) (string, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	value, ok := s.data[sURL]
	if !ok {
		return "", fmt.Errorf("не найдено данных для short url = %q", sURL)
	}
	return value.OrigURL, nil
}

func (s *MemStorage) IDExists(ctx context.Context, id string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, found := s.data[id]
	return found
}

func (s *MemStorage) Size() int {
	return len(s.data)
}

func (s *MemStorage) LastUUID() int {
	return s.lastUUID
}
