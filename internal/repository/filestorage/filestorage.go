package filestorage

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

type Storage struct {
	filePath string
	data     []model.StorageRecord
}

func NewStorage(filePath string) (*Storage, error) {
	return &Storage{filePath: filePath}, nil
}

func (f *Storage) LoadData(ctx context.Context, rcv any) error {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, rcv); err != nil {
		return err
	}

	return nil
}

func (f *Storage) SaveData() error {
	file, err := os.OpenFile(f.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(f.data)
}

func (f *Storage) Add(ctx context.Context, id, sURL, oURL string) error {
	f.data = append(f.data, model.StorageRecord{UUID: id, ShortURL: sURL, OrigURL: oURL})
	return f.SaveData()
}
