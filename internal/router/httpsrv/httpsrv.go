package httpsrv

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	hdlr "github.com/nk87rus/go-musthave-shortener/internal/handler"
)

type Server struct {
	addr string
	repo hdlr.Storage
}

func New(storage hdlr.Storage) *Server {
	return &Server{
		addr: "localhost:8080",
		repo: storage,
	}
}

func (s *Server) Run() error {
	r := chi.NewRouter()

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

	short, err := hdlr.CreateShortURL(string(body), s.repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u := url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   short,
	}

	response, err := u.MarshalBinary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(response); err != nil {
		log.Println(err.Error())
	}
}

func (s *Server) restoreURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("Метод %q не поддерживается. Допустим только %q", r.Method, http.MethodGet), http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	fullURL, err := hdlr.RestoreURL(id, s.repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
