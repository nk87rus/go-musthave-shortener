package simple

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	require.IsType(t, &Storage{}, NewStorage())
}

func TestStorageAdd(t *testing.T) {
	var (
		testKey   = "t"
		testValue = "a"
	)
	s := Storage{data: make(map[string]string)}
	require.Len(t, s.data, 0)
	s.Add(testKey, testValue)
	require.Len(t, s.data, 1)
	v, ok := s.data[testKey]
	require.True(t, ok)
	require.Equal(t, testValue, v)
}

func TestStorageGet(t *testing.T) {
	var (
		testKey     = "t"
		testValue   = "a"
		errNorFound = fmt.Errorf("не найдено данных")
	)
	testCases := []struct {
		name      string
		data      map[string]string
		wantError error
	}{
		{
			name:      "errNorFound",
			data:      map[string]string{testValue: testKey},
			wantError: errNorFound,
		},
		{
			name:      "Correct",
			data:      map[string]string{testKey: testValue},
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
	s := Storage{data: map[string]string{"a": "b"}}
	require.True(t, s.IDExists("a"))
	require.False(t, s.IDExists("x"))
}
