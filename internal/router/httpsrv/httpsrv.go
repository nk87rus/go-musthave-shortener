// Модуль httpsrv реализует HTTP сервер, обрабазывающий пользовательские запрос.
package httpsrv

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	hdlr "github.com/nk87rus/go-musthave-shortener/internal/handler"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
)

// Auditor - описывает интерфейс необходимых методов для ведения аудита/логирования обрабатываемых данных
//
//go:generate go run github.com/vektra/mockery/v2 --name=Auditor --inpackage --testonly
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
	addr      string
	cert, key string
	baseURL   *url.URL
	db        hdlr.Database
	handlers  Handlers
	audit     Auditor
	useTLS    bool
}

type JSONReqBody struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

func New(address, baseAddress string, useTLS bool, storage hdlr.Storage, db hdlr.Database) (*Server, error) {
	baseURL, err := url.Parse(baseAddress)
	if err != nil {
		return nil, fmt.Errorf("не корректный base address: %w", err)
	}

	newSrv := Server{
		addr:     address,
		baseURL:  baseURL,
		handlers: hdlr.InitHandlers(baseURL, storage),
		db:       db,
		useTLS:   useTLS,
	}

	return &newSrv, nil
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

	var srv = http.Server{Addr: s.addr, Handler: r}
	idleConnsClosed := make(chan struct{})

	go func(srvCtx context.Context) {
		<-srvCtx.Done()
		if err := srv.Shutdown(srvCtx); err != nil {
			log.Err(err)
		}
		close(idleConnsClosed)
	}(ctx)

	if s.useTLS {
		if err := s.MakeCerts(); err != nil {
			return err
		}
		defer func() {
			if s.cert != "" {
				os.Remove(s.cert)
			}

			if s.key != "" {
				os.Remove(s.key)
			}
		}()
		if err := srv.ListenAndServeTLS(s.cert, s.key); err != http.ErrServerClosed {
			return err
		}
	} else {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
	}

	<-idleConnsClosed
	return nil
}

// MakeCerts - создаёт сертификат и ключ
// разумеется, решение не приемлимо для прома :о)
//
// Returns:
//   - путь к файлу секртификата
//   - путь к файлу ключа
//   - ошибку, при возникновении
func (s *Server) MakeCerts() error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Shortener"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 1, 0),
		SubjectKeyId: []byte{2, 0, 2, 6, 3},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	certFile, crtErr := os.CreateTemp("", "crt-*.txt")
	if err != nil {
		return crtErr
	}
	defer certFile.Close()
	if _, crtWriteErr := certFile.Write(certPEM.Bytes()); crtWriteErr != nil {
		return crtWriteErr
	}
	s.cert = certFile.Name()

	keyFile, keyErr := os.CreateTemp("", "key-*.txt")
	if err != nil {
		return keyErr
	}
	defer keyFile.Close()
	if _, keyWriteErr := keyFile.Write(privateKeyPEM.Bytes()); keyWriteErr != nil {
		return keyWriteErr
	}
	s.key = keyFile.Name()

	return nil
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

	s.sendAudit(r.Context(), "shorten", string(body))
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

	s.sendAudit(r.Context(), "shorten", bodyData.URL)
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

	s.sendAudit(r.Context(), "follow", fullURL)
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

func (s *Server) sendAudit(ctx context.Context, action, urlData string) {
	if s.audit != nil {
		userID, _ := ctx.Value(model.CtxUserID).(string)
		aMsg := model.AuditMsg{
			Timestamp: time.Now().Unix(),
			Action:    action,
			UserID:    userID,
			URL:       urlData,
		}
		if err := s.audit.Notify(ctx, aMsg); err != nil {
			log.Err(err)
		}
	}
}
