package simple

import (
	"maps"
	"os"
	"testing"

	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorasge(t *testing.T) {
	var fp = "test"
	resultData, resultError := NewFileStorage(fp)
	require.Nil(t, resultError)
	require.IsType(t, &FileStorage{}, resultData)
	require.Equal(t, fp, resultData.filePath)
}

func TestLoadData(t *testing.T) {
	const tmpFilePtrn string = "ld*.json"
	t.Run("file_not_exist", func(t *testing.T) {
		s := FileStorage{}
		err := s.LoadData(nil)
		require.NoError(t, err)
	})

	t.Run("err_unmarshal", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), tmpFilePtrn)
		require.NoError(t, err)
		defer os.Remove(f.Name())

		s := FileStorage{filePath: f.Name()}
		require.Error(t, s.LoadData(nil))
	})

	t.Run("correct", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), tmpFilePtrn)
		require.NoError(t, err)
		defer os.Remove(f.Name())

		_, err = f.WriteString(`[
		{"uuid": "1", "short_url": "s1", "original_url": "o1"},
		{"uuid": "2", "short_url": "s2", "original_url": "o2"}
		]`)
		require.NoError(t, err)
		f.Close()

		fs := FileStorage{filePath: f.Name()}
		s := Storage{}
		require.Len(t, s.data, 0)
		require.NoError(t, fs.LoadData(&s))
		require.Len(t, s.data, 2)
		require.Equal(t, 2, s.lastUUID)
	})
}

func TestSaveData(t *testing.T) {
	const tmpFilePtrn string = "sd*.json"
	f, err := os.CreateTemp(os.TempDir(), tmpFilePtrn)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	stat, err := os.Stat(f.Name())
	require.NoError(t, err)
	require.Equal(t, int64(0), stat.Size())

	fs := FileStorage{filePath: f.Name()}
	data := map[int]model.StorageRecord{0: {UUID: "1", ShortURL: "s", OrigURL: "o"}}
	resultError := fs.SaveData(maps.Values(data))
	require.NoError(t, resultError)

	stat, err = os.Stat(f.Name())
	require.NoError(t, err)
	require.Greater(t, stat.Size(), int64(0))

	fData, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	require.Equal(t, `[{"uuid":"1","short_url":"s","original_url":"o"}]
`, string(fData))
}
