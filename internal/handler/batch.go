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
