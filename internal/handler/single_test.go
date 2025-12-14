package handler

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURL(t *testing.T) {
	errValue := fmt.Errorf("отсутсвует значение")
	testCases := []struct {
		name      string
		data      string
		mFunc     func(m *MockStorage)
		wantError error
	}{
		{
			name:      "errValue",
			wantError: errValue,
		},
		{
			name: "Correct",
			data: "test",
			mFunc: func(m *MockStorage) {
				m.On("IDExists", mock.Anything, mock.AnythingOfType("string")).Return(false)
				m.On("Add", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			sMock := NewMockStorage(t)
			if tc.mFunc != nil {
				tc.mFunc(sMock)
			}
			resultData, _, resultError := (&Handlers{repo: sMock, baseURL: &url.URL{}}).CreateShortURL(context.Background(), tc.data)
			if tc.wantError != nil {
				require.Empty(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
				require.NotEmpty(t, resultData)
			}
		})
	}
}

func TestRestoreURL(t *testing.T) {
	errEmptyID := fmt.Errorf("не указан идентификатор")
	testCases := []struct {
		name      string
		data      string
		mFunc     func(m *MockStorage)
		wantError error
	}{
		{
			name:      "errEmptyID",
			wantError: errEmptyID,
		},
		{
			name: "Correct",
			data: "test",
			mFunc: func(m *MockStorage) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("testURL", false, nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sMock := NewMockStorage(t)
			if tc.mFunc != nil {
				tc.mFunc(sMock)
			}

			resultData, _, resultError := (&Handlers{repo: sMock}).RestoreURL(context.Background(), tc.data)
			if tc.wantError != nil {
				require.Empty(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
				require.NotEmpty(t, resultData)
			}
		})
	}
}
