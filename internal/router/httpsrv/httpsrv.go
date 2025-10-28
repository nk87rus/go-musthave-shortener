package httpsrv

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	hdlr "github.com/nk87rus/go-musthave-shortener/internal/handler"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Handlers --inpackage --testonly
type Handlers interface {
	CreateShortURL(value string, repo hdlr.Storage) (string, error)
	RestoreURL(id string, repo hdlr.Storage) (string, error)
}

type Server struct {
	addr     string
	baseURL  *url.URL
	repo     hdlr.Storage
	handlers Handlers
}

func New(address, baseAddress string, storage hdlr.Storage) (*Server, error) {
	baseURL, err := url.Parse(baseAddress)
	if err != nil {
		return nil, fmt.Errorf("не корректный base address: %w", err)
	}
	return &Server{addr: address, baseURL: baseURL, repo: storage, handlers: &hdlr.Handlers{}}, nil
}

func (s *Server) Run() error {
	log.Info().Str("address", s.addr).Str("baseAddress", s.baseURL.String()).Msg("Запуск HTTP сервера")
	r := chi.NewRouter()
	r.Use(loggerMiddleware)

	r.Post("/", s.createShortURL)
	r.Get("/{id}", s.restoreURL)

	return http.ListenAndServe(s.addr, r)
}

func (s *Server) createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Метод %q не поддерживается. Допустим только %q", r.Method, http.MethodPost), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	short, err := s.handlers.CreateShortURL(string(body), s.repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newURL := *s.baseURL
	newURL.Path = short

	response, err := newURL.MarshalBinary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(response); err != nil {
		log.Err(err)
	}
}

func (s *Server) restoreURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Метод %q не поддерживается. Допустим только %q", r.Method, http.MethodGet), http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	fullURL, err := s.handlers.RestoreURL(id, s.repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
