package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/nk87rus/go-musthave-shortener/internal/repository"
)

func (h *Handlers) CreateShortURL(ctx context.Context, value string) (*url.URL, int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("отсутсвует значение для обработки")
	}

	strResult, err := h.GetRandomString(ctx)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err := h.repo.Add(ctx, strResult, value); err != nil {
		var dbErr *repository.DBError
		if errors.As(err, &dbErr) {
			return h.MakeShortNameURL(dbErr.Value), dbErr.HTTPResponseCode, nil
		} else {
			return nil, http.StatusBadRequest, err
		}
	}
	return h.MakeShortNameURL(strResult), http.StatusCreated, nil
}

func (h *Handlers) RestoreURL(ctx context.Context, id string) (string, bool, error) {
	if strings.TrimSpace(id) == "" {
		return "", false, fmt.Errorf("не указан идентификатор")
	}

	return h.repo.Get(ctx, id)
}
