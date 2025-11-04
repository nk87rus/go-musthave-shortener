package handler

import (
	"fmt"
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
				m.On("IDExists", mock.AnythingOfType("string")).Return(false)
				m.On("Add", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
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
			resultData, resultError := (&Handlers{}).CreateShortURL(tc.data, sMock)
			if tc.wantError != nil {
				require.Empty(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
				require.NotEmpty(t, resultData)
				require.Len(t, resultData, 8)
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
				m.On("Get", mock.AnythingOfType("string")).Return("testURL", nil)
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

			resultData, resultError := (&Handlers{}).RestoreURL(tc.data, sMock)
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
