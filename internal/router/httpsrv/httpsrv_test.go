package httpsrv

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	var errBaseURL = fmt.Errorf("не корректный base address")

	testCases := []struct {
		name       string
		addr       string
		baddr      string
		wantResult Server
		wantError  error
	}{
		{
			name:      "errBaseURL",
			baddr:     ":abc.xx",
			wantError: errBaseURL,
		},
		{
			name:  "Correct",
			addr:  "127.0.0.1:8080",
			baddr: "http://127.0.0.2:8090",
			wantResult: Server{
				addr:    "127.0.0.1:8080",
				baseURL: &url.URL{Scheme: "http", Host: "127.0.0.2:8090"},
				repo:    nil},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultData, resultError := New(tc.addr, tc.baddr, nil)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &Server{}, resultData)
				require.EqualValues(t, tc.wantResult, *resultData)
			}
		})
	}
}

func TestServerRun(t *testing.T) {
	s := &Server{}
	go s.Run()
}

// func TestServerCreateShortURL(t *testing.T) {
// 	var (
// 		errMethod  error = fmt.Errorf("не поддерживается")
// 		errReadAll error = fmt.Errorf("errReadAll")
// 		errHDLR    error = fmt.Errorf("errHDLR")
// 		errMarshal error = fmt.Errorf("errMarshal")
// 		errWrite   error = fmt.Errorf("errWrite")
// 	)
// 	testCases := []struct {
// 		name      string
// 		req       *http.Request
// 		wantError error
// 	}{
// 		{
// 			name:      "errMethod",
// 			req:       httptest.NewRequest(http.MethodPatch, "/", nil),
// 			wantError: errMethod,
// 		},
// 		{
// 			name:      "errReadAll",
// 			req:       httptest.NewRequest(http.MethodPost, "/", nil),
// 			wantError: errReadAll,
// 		},
// 		{
// 			name:      "errHDLR",
// 			req:       httptest.NewRequest(http.MethodPost, "/", nil),
// 			wantError: errHDLR,
// 		},
// 		{
// 			name:      "errMarshal",
// 			req:       httptest.NewRequest(http.MethodPost, "/", nil),
// 			wantError: errMarshal,
// 		},
// 		{
// 			name:      "errWrite",
// 			req:       httptest.NewRequest(http.MethodPost, "/", nil),
// 			wantError: errWrite,
// 		},
// 		{
// 			name:      "Correct",
// 			req:       httptest.NewRequest(http.MethodPost, "/", nil),
// 			wantError: nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			patchReadAll := monkey.Patch(io.ReadAll,
// 				func(io.Reader) ([]byte, error) {
// 					if errors.Is(tc.wantError, errReadAll) {
// 						return nil, tc.wantError
// 					}
// 					return []byte("test"), nil
// 				})
// 			defer patchReadAll.Unpatch()

// 			patchHCreateShortURL := monkey.Patch(handler.CreateShortURL,
// 				func(string, handler.Storage) (string, error) {
// 					if errors.Is(tc.wantError, errHDLR) {
// 						return "", tc.wantError
// 					}
// 					return "test1", nil
// 				})
// 			defer patchHCreateShortURL.Unpatch()

// 			patchMarshal := monkey.PatchInstanceMethod(reflect.TypeOf(&url.URL{}), "MarshalBinary",
// 				func(*url.URL) ([]byte, error) {
// 					if errors.Is(tc.wantError, errMarshal) {
// 						return nil, tc.wantError
// 					}
// 					return []byte("test2"), nil
// 				})
// 			defer patchMarshal.Unpatch()

// 			w := httptest.NewRecorder()

// 			if errors.Is(tc.wantError, errWrite) {
// 				patchWrite := monkey.PatchInstanceMethod(reflect.TypeOf(&httptest.ResponseRecorder{}), "Write",
// 					func(*httptest.ResponseRecorder, []byte) (int, error) {
// 						return 0, tc.wantError
// 					})
// 				defer patchWrite.Unpatch()
// 			}
// 			(&Server{}).createShortURL(w, tc.req)

// 			if tc.wantError != nil && !errors.Is(tc.wantError, errWrite) {
// 				require.Equal(t, http.StatusBadRequest, w.Code)
// 			} else {
// 				require.Equal(t, http.StatusCreated, w.Code)
// 				require.Equal(t, "text/plain", w.Header().Get("Content-Type"))
// 				require.Equal(t, "24", w.Header().Get("Content-Length"))
// 			}
// 		})
// 	}
// }

// func TestServerRestoreURL(t *testing.T) {
// 	var (
// 		errMethod error = fmt.Errorf("errMethod")
// 		errHDLR   error = fmt.Errorf("errHDLR")
// 	)
// 	testCases := []struct {
// 		name      string
// 		req       *http.Request
// 		wantError error
// 	}{
// 		{
// 			name:      "errMethod",
// 			req:       httptest.NewRequest(http.MethodPatch, "/", nil),
// 			wantError: errMethod,
// 		},
// 		{
// 			name:      "errHDLR",
// 			req:       httptest.NewRequest(http.MethodGet, "/123", nil),
// 			wantError: errHDLR,
// 		},
// 		{
// 			name:      "Correct",
// 			req:       httptest.NewRequest(http.MethodGet, "/123", nil),
// 			wantError: nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			patchHDLR := monkey.Patch(handler.RestoreURL,
// 				func(string, handler.Storage) (string, error) {
// 					if errors.Is(tc.wantError, errHDLR) {
// 						return "", tc.wantError
// 					}
// 					return "a", nil
// 				})
// 			defer patchHDLR.Unpatch()

// 			w := httptest.NewRecorder()
// 			(&Server{}).restoreURL(w, tc.req)

// 			if tc.wantError != nil {
// 				require.Equal(t, http.StatusBadRequest, w.Code)
// 			} else {
// 				require.Equal(t, http.StatusTemporaryRedirect, w.Code)
// 				require.Equal(t, "a", w.Header().Get("Location"))
// 			}
// 		})
// 	}
// }
