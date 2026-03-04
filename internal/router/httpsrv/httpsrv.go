// Модуль httpsrv реализует HTTP сервер, обрабазывающий пользовательские запрос.
package httpsrv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	hdlr "github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

// Auditor - описывает интерфейс необходимых методов для ведения аудита/логирования обрабатываемых данных
type Auditor interface {
	Notify(ctx context.Context, data model.AuditMsg) error
}

// Handlers - описывет набор методов, реализующих основной функционал приложения по созданию коротких URL и восстановлению из коротких URL их базовых значений
//
//go:generate go run github.com/vektra/mockery/v2 --name=Handlers --inpackage --testonly
type Handlers interface {
	CreateShortURL(ctx context.Context, value string) (*url.URL, int, error)
	RestoreURL(ctx context.Context, id string) (string, bool, error)
	CreateShortURLBatch(ctx context.Context, batch io.Reader) ([]byte, error)
	GetUsersURLs(ctx context.Context) ([]byte, error)
	DelURLs(ctx context.Context, data []string)
}

type Server struct {
	addr     string
	baseURL  *url.URL
	db       hdlr.Database
	handlers Handlers
	audit    Auditor
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

// EnableAudit - активирует функции аудита обрабатываемых данных
// 
// Args:
//   - filePath - путь к файлу для сохранения аудита
//   - urlPath  - URL для отправки данных аудита
//
// Допускается сохранение данных аудита как в файл так и на удалённый сервер, посредством отправки данных на указанный URL.
//
// Для сохранения данный в файл, необходимо указать путь к этому файлу.
//
// Так же для отправки данных на удалённый сервер, необходимо указать URL этого сервера.
//
// Если оставить пустым какой-либо ииз аргументов, то соответвующее этому типу хранилице данных не будет задействовано.
func (s *Server) EnableAudit(filePath, urlPath string) {
	log.Info().Str("filePath", filePath).Str("urlPath", urlPath).Msg("Подключение аудита запросов")
	defer log.Info().Str("filePath", filePath).Str("urlPath", urlPath).Msg("Подключение аудита запросов завершено")

	s.audit = InitAudit(filePath, urlPath)
}

func (s *Server) Run(ctx context.Context) error {
	log.Info().Str("address", s.addr).Str("baseAddress", s.baseURL.String()).Msg("Запуск HTTP сервера")
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Use(loggerMiddleware)
	r.Use(authMiddleware)

	r.Get("/{id}", s.restoreURL)
	r.Post("/api/shorten", s.createShortURLFromJSON)
	r.Post("/api/shorten/batch", s.createShortURLFromJSONBatch)
	r.Get("/api/user/urls", s.userURLs)
	r.Delete("/api/user/urls", s.delURLs)
	r.Get("/ping", s.pingDB)
	r.Post("/", s.createShortURL)

	return http.ListenAndServe(s.addr, r)
}

func (s *Server) createShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errMsg := fmt.Sprintf("метод %q не поддерживается. Допустим только %q", r.Method, http.MethodPost)
		fmt.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Err(err).Msg("createShortURL")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newURL, rCode, err := s.handlers.CreateShortURL(r.Context(), string(body))
	if err != nil {
		http.Error(w, err.Error(), rCode)
	}

	response, err := newURL.MarshalBinary()
	if err != nil {
		fmt.Println(err.Error())
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

	if s.audit != nil {
		userID, _ := r.Context().Value(model.CtxUserID).(string)
		aMsg := model.AuditMsg{
			Timestamp: time.Now().Unix(),
			Action:    "shorten",
			UserID:    userID,
			URL:       string(body),
		}
		if err := s.audit.Notify(r.Context(), aMsg); err != nil {
			log.Err(err)
		}
	}
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

	if s.audit != nil {
		userID, _ := r.Context().Value(model.CtxUserID).(string)
		aMsg := model.AuditMsg{
			Timestamp: time.Now().Unix(),
			Action:    "shorten",
			UserID:    userID,
			URL:       bodyData.URL,
		}
		if err := s.audit.Notify(r.Context(), aMsg); err != nil {
			log.Err(err)
		}
	}
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
	fullURL, isDeleted, err := s.handlers.RestoreURL(r.Context(), id)
	if err != nil {
		log.Err(err).Msg("restoreURL")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if isDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

	if s.audit != nil {
		userID, _ := r.Context().Value(model.CtxUserID).(string)
		aMsg := model.AuditMsg{
			Timestamp: time.Now().Unix(),
			Action:    "follow",
			UserID:    userID,
			URL:       fullURL,
		}
		if err := s.audit.Notify(r.Context(), aMsg); err != nil {
			log.Err(err)
		}
	}
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
		errMsg := fmt.Sprintf("метод %q не поддерживается. Допустим только %q", r.Method, http.MethodGet)
		log.Error().Msg(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	uid := r.Context().Value(model.CtxUserID)
	if uid == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// if uid == "" {
	// 	w.WriteHeader(http.StatusNoContent)
	// 	return
	// }

	result, err := s.handlers.GetUsersURLs(r.Context())
	if err != nil {
		fmt.Println(err.Error())
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

func (s *Server) delURLs(w http.ResponseWriter, r *http.Request) {
	var delList []string
	if err := json.NewDecoder(r.Body).Decode(&delList); err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(delList) == 0 {
		errMsg := "список URL для удаления пуст"
		log.Error().Msg(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	go s.handlers.DelURLs(r.Context(), delList)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

}
