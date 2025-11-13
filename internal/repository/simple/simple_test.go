package simple

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	var errLoadData = fmt.Errorf("errLoad")

	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errLoadData",
			wantError: errLoadData,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchLoadData := monkey.PatchInstanceMethod(reflect.TypeOf(&FileStorage{}), "LoadData",
				func(*FileStorage, any) error {
					if errors.Is(tc.wantError, errLoadData) {
						return tc.wantError
					}
					return nil
				})
			defer patchLoadData.Unpatch()

			resultData, resultError := NewStorage("")
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &Storage{}, resultData)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Run("err_file_unmarshal", func(t *testing.T) {
		s := Storage{}
		require.Error(t, json.Unmarshal(nil, &s))
	})
	t.Run("correct", func(t *testing.T) {
		data := []byte(`[
		{"uuid": "1", "short_url": "s1", "original_url": "o1"},
		{"uuid": "3a", "short_url": "s3", "original_url": "o3"},
		{"uuid": "2", "short_url": "s2", "original_url": "o2"}
		]`)
		s := Storage{}
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

	s := Storage{data: make(map[string]model.StorageRecord), extStorage: esMock}
	require.Len(t, s.data, 0)
	s.Add(testKey, testValue)
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
			s := Storage{data: tc.data}
			resultData, resultError := s.Get(testKey)
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
	s := Storage{data: map[string]model.StorageRecord{"a": {}}}
	require.True(t, s.IDExists("a"))
	require.False(t, s.IDExists("x"))
}
