package handler

import (
	"net/url"
	"testing"

	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInitHandlers(t *testing.T) {
	h := InitHandlers(nil, nil)
	require.IsType(t, &Handlers{}, h)
}

func TestMakeShortNameURL(t *testing.T) {
	h := &Handlers{baseURL: &url.URL{Scheme: "http", Host: "test.ru"}}
	resultData := h.MakeShortNameURL("sn")
	require.Equal(t, "http://test.ru/sn", resultData.String())
}

func TestGetRandomString(t *testing.T) {
	testCases := []struct {
		name      string
		mFunc     func(m *MockStorage)
		wantError error
	}{
		{
			name: "Correct",
			mFunc: func(m *MockStorage) {
				m.On("IDExists", mock.Anything, mock.AnythingOfType("string")).
					Return(false)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rMock := NewMockStorage(t)
			if tc.mFunc != nil {
				tc.mFunc(rMock)
			}
			h := &Handlers{repo: rMock}
			resultData, resultError := h.GetRandomString(t.Context())
			if tc.wantError != nil {
				require.Empty(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.NotEmpty(t, resultData)
				require.Nil(t, resultError)
			}

		})
	}
}

func BenchmarkGetRandomString(b *testing.B) {
	b.StopTimer()
	rMock := NewMockStorage(b)
	rMock.On("IDExists", mock.Anything, mock.AnythingOfType("string")).
		Return(false)
	h := &Handlers{repo: rMock}
	b.StartTimer()

	b.Run("GetRandomString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			h.GetRandomString(b.Context())
		}
	})
}
