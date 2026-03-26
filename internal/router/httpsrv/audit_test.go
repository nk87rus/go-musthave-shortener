package httpsrv

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWithFileAudit(t *testing.T) {
	a := &Audit{}
	require.Empty(t, a.subscribers)
	WithFileAudit("test")(a)
	require.Len(t, a.subscribers, 1)
}

func TestWithURLAudit(t *testing.T) {
	a := &Audit{}
	WithURLAudit("test")(a)
}

func TestFANOtify(t *testing.T) {
	tf, err := os.CreateTemp("", "fa-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tf.Name())
	fa := FileAudit{path: tf.Name()}
	fa.Notify(t.Context(), model.AuditMsg{})
}

func TestUANotify(t *testing.T) {
	var errReq = fmt.Errorf("audit response code")
	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errReq",
			wantError: errReq,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(
				http.HandlerFunc(
					func(rw http.ResponseWriter, req *http.Request) {
						if errors.Is(tc.wantError, errReq) {
							rw.WriteHeader(http.StatusBadRequest)
						} else {
							rw.WriteHeader(http.StatusOK)
						}
					},
				),
			)
			defer srv.Close()

			ua := URLAudit{path: srv.URL, client: resty.New()}
			resultError := ua.Notify(t.Context(), model.AuditMsg{})
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
			}
		})
	}
}

func TestAuditNotify(t *testing.T) {
	var errNotify = fmt.Errorf("errNotify")

	sMock1 := NewMockSubscriber(t)
	sMock1.On("Notify", mock.Anything, mock.Anything).Return(errNotify)
	sMock2 := NewMockSubscriber(t)
	sMock2.On("Notify", mock.Anything, mock.Anything).Return(nil)

	a := Audit{subscribers: []Subscriber{sMock1, sMock2}}
	resultError := a.Notify(t.Context(), model.AuditMsg{})
	require.Error(t, resultError)
}
