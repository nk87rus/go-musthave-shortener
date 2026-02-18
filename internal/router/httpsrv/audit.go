package httpsrv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/rs/zerolog/log"
)

type Subscriber interface {
	Notify(ctx context.Context, data model.AuditMsg) error
}

type Audit struct {
	mu          sync.RWMutex
	subscribers []Subscriber
}

func InitAudit(filePath, urlPath string) *Audit {
	newAudit := &Audit{}
	WithFileAudit(filePath)(newAudit)
	WithURLAudit(urlPath)(newAudit)
	return newAudit
}

func WithFileAudit(path string) func(*Audit) {
	return func(a *Audit) {
		if path != "" {
			a.mu.Lock()
			defer a.mu.Unlock()
			a.subscribers = append(a.subscribers, &FileAudit{path: path})
		}
	}
}

func WithURLAudit(path string) func(*Audit) {
	return func(a *Audit) {
		if path != "" {
			a.mu.Lock()
			defer a.mu.Unlock()
			a.subscribers = append(a.subscribers, &URLAudit{path})
		}
	}
}

type FileAudit struct {
	path string
}

func (fa *FileAudit) Notify(ctx context.Context, data model.AuditMsg) error {
	log.Debug().Msgf("Запись аудита в %s", fa.path)
	defer log.Debug().Msgf("Запись аудита в %s завершена", fa.path)
	f, err := os.OpenFile(fa.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(data); err != nil {
		f.Close()
		return err
	}

	return nil
}

type URLAudit struct {
	path string
}

func (ua *URLAudit) Notify(ctx context.Context, data model.AuditMsg) error {
	log.Debug().Msgf("Запись аудита в %s", ua.path)
	defer log.Debug().Msgf("Запись аудита в %s завершена", ua.path)
	client := resty.New()
	resp, err := client.R().SetContext(ctx).SetBody(data).Post(ua.path)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("audtt response code: %d", resp.StatusCode())
	}

	return nil
}

func (a *Audit) Notify(ctx context.Context, data model.AuditMsg) error {
	var errPool []error = make([]error, len(a.subscribers))
	for _, s := range a.subscribers {
		if err := s.Notify(ctx, data); err != nil {
			errPool = append(errPool, err)
			log.Err(err).Msgf("Ошибка при записи аудита")
			continue
		}
	}
	return errors.Join(errPool...)
}
