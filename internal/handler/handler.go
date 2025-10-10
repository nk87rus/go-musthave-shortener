package handler

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Storage interface {
	Add(id, value string)
	Get(id string) (string, error)
	IdExists(id string) bool
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func CreateShortURL(value string, repo Storage) (string, error) {
	const resultLength = 8

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("отсутсвует значение для обработки")
	}

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	var strResult string
	for len(strResult) == 0 && !repo.IdExists(strResult) {
		result := make([]byte, resultLength)
		for i := range result {
			result[i] = charset[seededRand.Intn(len(charset))]
		}
		strResult = string(result)
	}

	repo.Add(strResult, value)
	return strResult, nil
}

func RestoreURL(id string, repo Storage) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("не указан идентификатор")
	}

	return repo.Get(id)
}
