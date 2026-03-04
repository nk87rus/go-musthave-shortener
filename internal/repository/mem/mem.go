// Модуль memstorage реализует функцонал хранения данных в оперативной памяти
package memstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strconv"
	"sync"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/rs/zerolog/log"
)

// ExtStorage - описывает интерфейс методов, необходимых для взаимодействия с внешними хранилищами
//
//go:generate go run github.com/vektra/mockery/v2 --name=ExtStorage --inpackage --testonly
type ExtStorage interface {
	LoadData(ctx context.Context, rcv any) error
	Add(ctx context.Context, id, sURL, oURL string) error
	AddBatch(ctx context.Context, data iter.Seq[model.StorageRecord]) error
	DelURLs(ctx context.Context, uid string, urls []string) error
}

// MemStorage - структура хранилища
type MemStorage struct {
	m          sync.RWMutex
	data       map[string]model.StorageRecord
	lastUUID   int        // последний добавленный индентификатор записи
	extStorage ExtStorage // внешнее хранилище
}

// NewStorage - инициализирует новой хранилище в оперативной памяти
//
// Args:
//   - extStorasge - внешнее хранилище
func NewStorage(ctx context.Context, extStorage ExtStorage) (*MemStorage, error) {
	var newStorage = MemStorage{
		lastUUID:   0,
		data:       make(map[string]model.StorageRecord),
		extStorage: extStorage,
	}

	if err := newStorage.extStorage.LoadData(ctx, &newStorage); err != nil {
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

// Add - добавляет запись в хранилище
//
// Args:
//   - id - идентификатор записи
//   - sURL - короткий URL
//   - oURL - исходный URL
func (s *MemStorage) Add(ctx context.Context, sURL, oURL string) error {
	userID, ok := ctx.Value(model.CtxUserID).(string)
	if !ok {
		return fmt.Errorf("не корректный тип userID (%T)", ctx.Value(model.CtxUserID))
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err := s.extStorage.Add(ctx, strconv.Itoa(s.lastUUID+1), sURL, oURL); err != nil {
		return err
	}
	s.lastUUID++
	s.data[sURL] = model.StorageRecord{UUID: strconv.Itoa(s.lastUUID), ShortURL: sURL, OrigURL: oURL, UserID: userID}
	// add to aeternal storage
	return nil
}

// AddBatch - ддобавлет набор записей в хранилище
//
// Args:
//   - data - список добавляемых записей
func (s *MemStorage) AddBatch(ctx context.Context, data *[]model.StorageRecord) error {
	userID, ok := ctx.Value(model.CtxUserID).(string)
	if !ok {
		return fmt.Errorf("не корректный тип userID (%T)", ctx.Value(model.CtxUserID))
	}

	s.m.Lock()
	defer s.m.Unlock()

	var esData = make([]model.StorageRecord, 0, len(*data))
	for _, rec := range *data {
		s.lastUUID++
		rec.UUID = strconv.Itoa(s.lastUUID)
		rec.UserID = userID
		s.data[rec.ShortURL] = rec
		esData = append(esData, rec)
	}

	return s.extStorage.AddBatch(ctx, slices.Values(esData))
}

// Get - извлекает из хранилища базовый URL по его короткому представлению
func (s *MemStorage) Get(ctx context.Context, sURL string) (string, bool, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	value, ok := s.data[sURL]
	if !ok {
		return "", false, fmt.Errorf("не найдено данных для short url = %q", sURL)
	}
	return value.OrigURL, value.DeletedFlag, nil
}

// IDExists - проверяет существование идентификатора в хранилище
//
// Args:
//   - id - проверяемый идентификатор
func (s *MemStorage) IDExists(ctx context.Context, id string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, found := s.data[id]
	return found
}

// Size - возвращает количество записей в хранилище
func (s *MemStorage) Size() int {
	return len(s.data)
}

// LastUUID - возвращает последний добавленный в хранилище идентификатор
func (s *MemStorage) LastUUID() int {
	return s.lastUUID
}

// GetUsersURLs - возвращает набор записей для определённого пользователя.
//
// Идентификатор пользователя извлекается из конткеста по ключу `model.CtxUserID`
func (s *MemStorage) GetUsersURLs(ctx context.Context) ([]model.StorageRecord, error) {
	userID, ok := ctx.Value(model.CtxUserID).(string)
	if !ok {
		return nil, fmt.Errorf("не корректный тип userID (%T)", ctx.Value(model.CtxUserID))
	}

	// if userID == "" {
	// 	return nil, errors.New("не задан ID  пользователя")
	// }

	s.m.RLock()
	defer s.m.RUnlock()
	var result []model.StorageRecord
	for _, v := range s.data {
		if v.UserID == userID {
			result = append(result, v)
		}
	}

	return result, nil
}

// DelURLs - удаляет записи из хранилища
//
// Args:
//   - uid - идентификатор пользователя-владельца записей
//   - urls - список удаляемых записей
func (s *MemStorage) DelURLs(ctx context.Context, uid string, urls []string) {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		s.m.Lock()
		defer s.m.Unlock()
		for _, sURL := range urls {
			if v, found := s.data[sURL]; found {
				if v.UserID == uid {
					v.DeletedFlag = true
					s.data[sURL] = v
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.extStorage.DelURLs(ctx, uid, urls); err != nil {
			log.Err(err).Msg("DelURLs: extStoerage error")
		}
	}()

	wg.Wait()
}
