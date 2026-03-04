// Модуль filestorage реализует функцонал хранения данных в виде фаqла на локальной файловой системе
package filestorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"slices"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/rs/zerolog/log"
)

// FileStorage - структура файлового хранилища
type FileStorage struct {
	filePath string // путь к файлу хранилища
}

// NewStorage - инициирует новой фаловое хранилище
//
// Args:
//   - filePath - путь к файлу хранилища
func NewStorage(filePath string) (*FileStorage, error) {
	return &FileStorage{filePath: filePath}, nil
}

// LoadData - загружает данный из файла-хранилища в указанный приёмник
//
// Args:
//   - rcv - приёмник данных из файла-хранилища
func (f *FileStorage) LoadData(ctx context.Context, rcv any) error {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, rcv); err != nil {
			return err
		}
	} else {
		log.Warn().Str("file", f.filePath).Msg("Файл хранилища пуст")
	}
	return nil
}

// SaveData - записывает данные в файл-хранилище
//
// Args:
//   - data - данные для записи
func (f *FileStorage) SaveData(data []model.StorageRecord) error {
	file, err := os.OpenFile(f.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(data)
}

// Add - добавляет запись в хранилище
//
// Args:
//   - id - идентификатор записи
//   - sURL - короткий URL
//   - oURL - исходный URL
func (f *FileStorage) Add(ctx context.Context, id, sURL, oURL string) error {
	var fData []model.StorageRecord
	if err := f.LoadData(ctx, &fData); err != nil {
		return err
	}
	userID, ok := ctx.Value(model.CtxUserID).(string)
	if !ok {
		return fmt.Errorf("не корректный тип userID (%T)", ctx.Value(model.CtxUserID))
	}

	fData = append(fData, model.StorageRecord{UUID: id, ShortURL: sURL, OrigURL: oURL, UserID: userID})
	return f.SaveData(fData)
}

// AddBatch - ддобавлет набор записей в хранилище
//
// Args:
//   - data - итератор по добавляемым записям
func (f *FileStorage) AddBatch(ctx context.Context, data iter.Seq[model.StorageRecord]) error {
	var fData []model.StorageRecord
	if err := f.LoadData(ctx, &fData); err != nil {
		return err
	}

	for rec := range data {
		fData = append(fData, rec)
	}

	return f.SaveData(fData)
}

// DelURLs - удаляет записи из хранилища
//
// Args:
//   - uid - идентификатор пользователя-владельца записей
//   - urls - список удаляемых записей
func (f *FileStorage) DelURLs(ctx context.Context, uid string, urls []string) error {
	var fData []model.StorageRecord
	if err := f.LoadData(ctx, &fData); err != nil {
		return err
	}

	for i := 0; i < len(fData); i++ {
		if fData[i].UserID == uid {
			if slices.Contains(urls, fData[i].ShortURL) {
				log.Debug().Str("short_url", fData[i].ShortURL).Str("uid", uid).Msg("mrk as deleted")
				fData[i].DeletedFlag = true
			}
		}
	}

	return f.SaveData(fData)

}
