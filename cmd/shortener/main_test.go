package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetValue(t *testing.T) {
	testCases := []struct {
		name       string
		data       string
		wantResult string
	}{
		{
			name:       "Empty",
			wantResult: "N/A",
		},
		{
			name:       "OnlyWhiteSpace",
			data:       "   ",
			wantResult: "N/A",
		},
		{
			name:       "Correct",
			data:       "test",
			wantResult: "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, getValue(tc.data), tc.wantResult)
		})
	}
}
