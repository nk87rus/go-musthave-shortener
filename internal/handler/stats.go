package handler

import (
	"context"
	"encoding/json"
)

func (h Handlers) Stats(ctx context.Context) ([]byte, error) {
	rawData, err := h.repo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(rawData)
}
