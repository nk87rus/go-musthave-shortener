package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

type ReqBatchItem struct {
	ID      string `json:"correlation_id"`
	OrigURL string `json:"original_url"`
}

type RespBatchItem struct {
	ID      string `json:"correlation_id"`
	SortURL string `json:"short_url"`
}

type RespUsersURLs struct {
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url" `
}

func (h *Handlers) CreateShortURLBatch(ctx context.Context, batch io.Reader) ([]byte, error) {
	var batchData []ReqBatchItem
	if err := json.NewDecoder(batch).Decode(&batchData); err != nil {
		return nil, err
	}

	if len(batchData) == 0 {
		return nil, fmt.Errorf("отсутсвует значение для обработки")
	}

	var (
		responseBatch = make([]RespBatchItem, 0, len(batchData))
		repoBatch     = make([]model.StorageRecord, 0, len(batchData))
	)
	for _, itm := range batchData {
		short, err := h.GetRandomString(ctx)
		if err != nil {
			return nil, err
		}
		responseBatch = append(responseBatch, RespBatchItem{ID: itm.ID, SortURL: h.MakeShortNameURL(short).String()})
		repoBatch = append(repoBatch, model.StorageRecord{ShortURL: short, OrigURL: itm.OrigURL})
	}

	if err := h.repo.AddBatch(ctx, &repoBatch); err != nil {
		return nil, err
	}

	return json.Marshal(responseBatch)
}

func (h *Handlers) GetUsersURLs(ctx context.Context) ([]byte, error) {
	data, err := h.repo.GetUsersURLs(ctx)
	if err != nil {
		return nil, err
	}
	var resultData = make([]RespUsersURLs, 0, len(data))
	for _, sr := range data {
		resultData = append(resultData, RespUsersURLs{OrigURL: sr.OrigURL, ShortURL: h.MakeShortNameURL(sr.ShortURL).String()})
	}

	if len(resultData) == 0 {
		return nil, nil
	}

	return json.Marshal(resultData)
}
