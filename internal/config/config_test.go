package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	os.Args = []string{"test", "-a", "127.0.0.1"}
	resultData := InitConfig()

	require.IsType(t, &ConfigData{}, resultData)
	require.Equal(t, "127.0.0.1", resultData.Addr)
	require.Equal(t, "http://127.0.0.1", resultData.BaseAddr)
}
