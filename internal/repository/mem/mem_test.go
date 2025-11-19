package memstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	var errLoadData = fmt.Errorf("errLoad")

	testCases := []struct {
		name      string
		mFunc     func(m *MockExtStorage)
		wantError error
	}{
		{
			name: "errLoadData",
			mFunc: func(m *MockExtStorage) {
				m.On("LoadData", mock.Anything).Return(errLoadData)
			},
			wantError: errLoadData,
		},
		{
			name: "Correct",
			mFunc: func(m *MockExtStorage) {
				m.On("LoadData", mock.Anything).Return(nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			esMock := NewMockExtStorage(t)
			if tc.mFunc != nil {
				tc.mFunc(esMock)
			}
			resultData, resultError := NewStorage(esMock)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &MemStorage{}, resultData)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Run("err_file_unmarshal", func(t *testing.T) {
		s := MemStorage{}
		require.Error(t, json.Unmarshal(nil, &s))
	})
	t.Run("correct", func(t *testing.T) {
		data := []byte(`[
		{"uuid": "1", "short_url": "s1", "original_url": "o1"},
		{"uuid": "3a", "short_url": "s3", "original_url": "o3"},
		{"uuid": "2", "short_url": "s2", "original_url": "o2"}
		]`)
		s := MemStorage{}
		require.NoError(t, json.Unmarshal(data, &s))
		require.Len(t, s.data, 2)
		require.Equal(t, 2, s.lastUUID)
	})
}

func TestStorageAdd(t *testing.T) {
	var (
		testKey   = "t"
		testValue = "a"
	)

	esMock := NewMockExtStorage(t)
	esMock.On("SaveData", mock.Anything).Return(nil)

	s := MemStorage{data: make(map[string]model.StorageRecord), extStorage: esMock}
	require.Len(t, s.data, 0)
	s.Add(context.Background(), testKey, testValue)
	require.Len(t, s.data, 1)

	v, ok := s.data[testKey]
	require.True(t, ok)
	require.Equal(t, testValue, v.OrigURL)

}

func TestStorageGet(t *testing.T) {
	var (
		testKey     = "t"
		testValue   = "a"
		errNorFound = fmt.Errorf("не найдено данных")
	)
	testCases := []struct {
		name      string
		data      map[string]model.StorageRecord
		wantError error
	}{
		{
			name:      "errNorFound",
			data:      map[string]model.StorageRecord{},
			wantError: errNorFound,
		},
		{
			name:      "Correct",
			data:      map[string]model.StorageRecord{testKey: {OrigURL: testValue}},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := MemStorage{data: tc.data}
			resultData, resultError := s.Get(context.Background(), testKey)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Empty(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.Equal(t, testValue, resultData)
			}
		})
	}
}

func TestStorageIDExists(t *testing.T) {
	s := MemStorage{data: map[string]model.StorageRecord{"a": {}}}
	require.True(t, s.IDExists(context.Background(), "a"))
	require.False(t, s.IDExists(context.Background(), "x"))
}
