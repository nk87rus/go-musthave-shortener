package httpsrv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	hdlr "github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Handlers --inpackage --testonly
type Handlers interface {
	CreateShortURL(ctx context.Context, value string) (*url.URL, int, error)
	RestoreURL(ctx context.Context, id string) (string, error)
	CreateShortURLBatch(ctx context.Context, batch io.Reader) ([]byte, error)
	GetUsersURLs(ctx context.Context) ([]byte, error)
}

type Server struct {
	addr     string
	baseURL  *url.URL
	db       hdlr.Database
	handlers Handlers
}

type JSONReqBody struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

func New(address, baseAddress string, storage hdlr.Storage, db hdlr.Database) (*Server, error) {
	baseURL, err := url.Parse(baseAddress)
	if err != nil {
		return nil, fmt.Errorf("не корректный base address: %w", err)
	}
	return &Server{addr: address, baseURL: baseURL, handlers: hdlr.InitHandlers(baseURL, storage), db: db}, nil
}

func (s *Server) Run(ctx context.Context) error {
	log.Info().Str("address", s.addr).Str("baseAddress", s.baseURL.String()).Msg("Запуск HTTP сервера")
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Use(loggerMiddleware)
	r.Use(authMiddleware)

	r.Post("/", s.createShortURL)
	r.Get("/{id}", s.restoreURL)
	r.Post("/api/shorten", s.createShortURLFromJSON)
	r.Post("/api/shorten/batch", s.createShortURLFromJSONBatch)
	r.Get("/api/user/urls", s.userURLs)
	r.Get("/ping", s.pingDB)

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

	newURL, rCode, err := s.handlers.CreateShortURL(r.Context(), string(body))
	if err != nil {
		http.Error(w, err.Error(), rCode)
	}

	response, err := newURL.MarshalBinary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(rCode)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))

	if _, err := w.Write(response); err != nil {
		log.Err(err)
	}
	log.Debug().Str("source URL", string(body)).Str("shortenURL", newURL.String()).Msg("сокращённый URL  успешно сформирован")
}

func (s Server) createShortURLFromJSON(w http.ResponseWriter, r *http.Request) {
	var bodyData JSONReqBody
	if err := json.NewDecoder(r.Body).Decode(&bodyData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newURL, rCode, err := s.handlers.CreateShortURL(r.Context(), bodyData.URL)
	if err != nil {
		http.Error(w, err.Error(), rCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rCode)

	var respData = JSONResponse{Result: newURL.String()}
	if err := json.NewEncoder(w).Encode(respData); err != nil {
		log.Err(err).Msg("ошибка маршаллинга ответа")
		return
	}
	log.Debug().Str("source URL", bodyData.URL).Str("shortenURL", respData.Result).Msg("сокращённый URL  успешно сформирован")
}

func (s Server) createShortURLFromJSONBatch(w http.ResponseWriter, r *http.Request) {
	shortBatch, err := s.handlers.CreateShortURLBatch(r.Context(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(shortBatch); err != nil {
		log.Err(err)
	}

	log.Debug().Msg("пакет сокращённых URL успешно сформирован")
}

func (s *Server) restoreURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("метод %q не поддерживается. Допустим только %q", r.Method, http.MethodGet), http.StatusBadRequest)
		return
	}
	id := r.PathValue("id")
	fullURL, err := s.handlers.RestoreURL(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) pingDB(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		http.Error(w, "подключение к БД не инициализировано", http.StatusInternalServerError)
		return
	}

	if err := s.db.Ping(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) userURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("метод %q не поддерживается. Допустим только %q", r.Method, http.MethodGet), http.StatusBadRequest)
		return
	}

	uid := r.Context().Value(model.CtxUserID)
	if uid == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	result, err := s.handlers.GetUsersURLs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if result == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(result); err != nil {
		log.Err(err)
	}

	log.Debug().Int("count", len(result)).Msg("список пользовательсиких URL успешно передан")
}
