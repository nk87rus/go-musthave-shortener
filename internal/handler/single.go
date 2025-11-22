package handler

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (h *Handlers) CreateShortURL(ctx context.Context, value string) (*url.URL, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("отсутсвует значение для обработки")
	}

	strResult, err := h.GetRandomString(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Add(ctx, strResult, value); err != nil {
		return nil, err
	}

	return h.MakeShortNameURL(strResult), nil
}

func (h *Handlers) RestoreURL(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("не указан идентификатор")
	}

	return h.repo.Get(ctx, id)
}
