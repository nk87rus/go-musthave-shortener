package handler

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Storage --inpackage --testonly
type Storage interface {
	Add(sURL, oURL string) error
	Get(sURL string) (string, error)
	IDExists(sURL string) bool
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Handlers struct{}

func (h *Handlers) CreateShortURL(value string, repo Storage) (string, error) {
	const resultLength = 8

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("отсутсвует значение для обработки")
	}

	var strResult string
	for len(strResult) == 0 || repo.IDExists(strResult) {
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

	if err := repo.Add(strResult, value); err != nil {
		return "", err
	}
	return strResult, nil
}

func (h *Handlers) RestoreURL(id string, repo Storage) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("не указан идентификатор")
	}

	return repo.Get(id)
}
