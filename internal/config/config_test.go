package config

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		resultData, resultError := InitConfig([]string{"test"})

		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, defaultAddr, resultData.Addr)
		require.Equal(t, "http://"+defaultAddr, resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("Flags - all", func(t *testing.T) {
		resultData, resultError := InitConfig([]string{"test", "-a", "127.0.0.1", "-b", "http://127.0.0.2", "-d", "TEST_DSN"})

		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.1", resultData.Addr)
		require.Equal(t, "http://127.0.0.2", resultData.BaseAddr)
		require.Equal(t, "TEST_DSN", resultData.DBDSN)
	})

	t.Run("Flags - address only", func(t *testing.T) {
		resultData, resultError := InitConfig([]string{"test", "-a", "127.0.0.1"})

		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.1", resultData.Addr)
		require.Equal(t, "http://127.0.0.1", resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("Flags - base url only", func(t *testing.T) {
		resultData, resultError := InitConfig([]string{"test", "-b", "http://127.0.0.2"})

		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, defaultAddr, resultData.Addr)
		require.Equal(t, "http://127.0.0.2", resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("ENVS ", func(t *testing.T) {
		t.Setenv("SERVER_ADDRESS", "127.0.0.3")
		t.Setenv("BASE_URL", "http://127.0.0.4")
		t.Setenv("DATABASE_DSN", "TEST_DSN")

		resultData, resultError := InitConfig([]string{"test"})
		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.3", resultData.Addr)
		require.Equal(t, "http://127.0.0.4", resultData.BaseAddr)
		require.Equal(t, "TEST_DSN", resultData.DBDSN)
	})

	t.Run("ENVS - addr only", func(t *testing.T) {
		t.Setenv("SERVER_ADDRESS", "127.0.0.3")

		resultData, resultError := InitConfig([]string{"test"})
		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.3", resultData.Addr)
		require.Equal(t, "http://127.0.0.3", resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("ENVS - base url only", func(t *testing.T) {
		t.Setenv("BASE_URL", "http://127.0.0.4")

		resultData, resultError := InitConfig([]string{"test"})
		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, defaultAddr, resultData.Addr)
		require.Equal(t, "http://127.0.0.4", resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("ENV - base url; Flag - addr", func(t *testing.T) {
		t.Setenv("BASE_URL", "http://127.0.0.4")

		resultData, resultError := InitConfig([]string{"test", "-a", "127.0.0.1"})
		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.1", resultData.Addr)
		require.Equal(t, "http://127.0.0.4", resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("ENV - addr; Flag - base url", func(t *testing.T) {
		t.Setenv("SERVER_ADDRESS", "127.0.0.3")

		resultData, resultError := InitConfig([]string{"test", "-b", "http://127.0.0.2"})
		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.3", resultData.Addr)
		require.Equal(t, "http://127.0.0.2", resultData.BaseAddr)
		require.Empty(t, resultData.DBDSN)
	})

	t.Run("ENV - addr; Flag - base url; File - dsn", func(t *testing.T) {
		t.Setenv("SERVER_ADDRESS", "127.0.0.3")
		t.Setenv("CONFIG", "test")

		patchRead := monkey.Patch(os.ReadFile,
			func(string) ([]byte, error) {
				return []byte(`{"server_address": "localhost:8080", "base_url": "http://localhost", "file_storage_path": "/path/to/file.db", "database_dsn": "DSN_FROM_FILE", "enable_https": false}`), nil
			})
		defer patchRead.Unpatch()

		resultData, resultError := InitConfig([]string{"test", "-b", "http://127.0.0.2"})
		require.Nil(t, resultError)
		require.IsType(t, &ConfigData{}, resultData)
		require.Equal(t, "127.0.0.3", resultData.Addr)
		require.Equal(t, "http://127.0.0.2", resultData.BaseAddr)
		require.Equal(t, "DSN_FROM_FILE", resultData.DBDSN)
	})
}

func TestCheckBaseURL(t *testing.T) {
	testCases := []struct {
		name       string
		data       string
		wantResult string
	}{
		{
			name:       "NoData",
			wantResult: "http://123",
		},
		{
			name:       "WithData",
			data:       "http://321",
			wantResult: "http://321",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cd := ConfigData{Addr: "123", BaseAddr: tc.data}
			cd.CheckBaseURL()
			require.Equal(t, tc.wantResult, cd.BaseAddr)
		})
	}
}

func TestCheckField(t *testing.T) {
	var (
		s1 = ""
		s2 = "a"
	)
	checkStringField(&s1, &s2)
	require.Equal(t, s1, s2)
}

func TestReadConfigFile(t *testing.T) {
	var errReadFile = fmt.Errorf("errReadFile")
	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errReadFile",
			wantError: errReadFile,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchRead := monkey.Patch(os.ReadFile,
				func(string) ([]byte, error) {
					if errors.Is(tc.wantError, errReadFile) {
						return nil, tc.wantError
					}
					return []byte(`{"server_address": "localhost:8080", "base_url": "http://localhost", "file_storage_path": "/path/to/file.db", "database_dsn": "", "enable_https": true}`), nil
				})
			defer patchRead.Unpatch()

			p := Parser{}
			resultError := p.ReadConfigFile("")
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Empty(t, p.fileData)
			} else {
				require.Nil(t, resultError)
				require.NotEmpty(t, p.fileData)
			}

		})
	}
}
