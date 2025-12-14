package handler

import (
	"context"
	"crypto/rand"
	"math/big"
	"net/url"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Storage --inpackage --testonly
type Storage interface {
	Add(ctx context.Context, sURL, oURL string) error
	AddBatch(ctx context.Context, data *[]model.StorageRecord) error
	Get(ctx context.Context, sURL string) (string, bool, error)
	GetUsersURLs(ctx context.Context) ([]model.StorageRecord, error)
	IDExists(ctx context.Context, sURL string) bool
	DelURLs(ctx context.Context, uid string, urls []string)
}

type Database interface {
	Ping(ctx context.Context) error
}

const (
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	resultLength = 8
)

type Handlers struct {
	baseURL *url.URL
	repo    Storage
}

func InitHandlers(baseURL *url.URL, repo Storage) *Handlers {
	return &Handlers{baseURL: baseURL, repo: repo}
}

func (h Handlers) MakeShortNameURL(shortName string) *url.URL {
	newURL := *h.baseURL
	newURL.Path = shortName

	return &newURL
}

func (h *Handlers) GetRandomString(ctx context.Context) (string, error) {
	var strResult string
	for len(strResult) == 0 || h.repo.IDExists(ctx, strResult) {
		result := make([]byte, resultLength)
		for i := range result {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			result[i] = charset[n.Int64()]
			strResult = string(result)
		}
	}

	return strResult, nil
}
