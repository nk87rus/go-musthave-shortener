package simple

import (
	"encoding/json"
	"errors"
	"iter"
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

type FileStorage struct {
	filePath string
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	return &FileStorage{filePath: filePath}, nil
}

func (f *FileStorage) LoadData(rcv any) error {
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

func (f *FileStorage) SaveData(data iter.Seq[model.StorageRecord]) error {
	var tmpData []model.StorageRecord = make([]model.StorageRecord, 0)
	for rec := range data {
		tmpData = append(tmpData, rec)
	}

	file, err := os.OpenFile(f.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(tmpData)
}
