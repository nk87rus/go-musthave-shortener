package jwtproc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("WithENV", func(t *testing.T) {
		t.Setenv("KEY", "123")
		resultData := New()
		require.IsType(t, &JWTProc{}, resultData)
		require.Equal(t, "123", resultData.key)
	})
	t.Run("WithOUTENV", func(t *testing.T) {
		resultData := New()
		require.IsType(t, &JWTProc{}, resultData)
		require.Equal(t, "super_secret_key", resultData.key)
	})
}

func TestTokenExpiration(t *testing.T) {
	require.Equal(t, tokenExp, (New()).TokenExpiration())
}

func TestMakeToken(t *testing.T) {
	jp := New()
	resultData, resultError := jp.MakeJWT()
	require.NotEmpty(t, resultData)
	require.Nil(t, resultError)
}
