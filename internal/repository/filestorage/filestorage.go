package filestorage

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"os"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

type Storage struct {
	filePath string
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

func (f *Storage) SaveData(data []model.StorageRecord) error {
	file, err := os.OpenFile(f.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(data)
}

func (f *Storage) Add(ctx context.Context, id, sURL, oURL string) error {
	var fData []model.StorageRecord
	if err := f.LoadData(ctx, &fData); err != nil {
		return err
	}

	fData = append(fData, model.StorageRecord{UUID: id, ShortURL: sURL, OrigURL: oURL})
	return f.SaveData(fData)
}

func (f *Storage) AddBatch(ctx context.Context, data iter.Seq[model.StorageRecord]) error {
	var fData []model.StorageRecord
	if err := f.LoadData(ctx, &fData); err != nil {
		return err
	}

	for rec :=  range data {
		fData = append(fData, rec)
	}

	return f.SaveData(fData)
}
