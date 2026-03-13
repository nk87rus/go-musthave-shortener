package filestorage

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/model"

	"github.com/stretchr/testify/require"
)

func TestNewFileStorasge(t *testing.T) {
	var fp = "test"
	resultData, resultError := NewStorage(fp)
	require.Nil(t, resultError)
	require.IsType(t, &FileStorage{}, resultData)
	require.Equal(t, fp, resultData.filePath)
}

func TestLoadData(t *testing.T) {
	const tmpFilePtrn string = "ld*.json"
	t.Run("file_not_exist", func(t *testing.T) {
		s := FileStorage{}
		err := s.LoadData(t.Context(), nil)
		require.NoError(t, err)
	})

	t.Run("err_unmarshal", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), tmpFilePtrn)
		require.NoError(t, err)
		defer func() {
			if errRemoveFile := os.Remove(f.Name()); errRemoveFile != nil {
				t.Log(errRemoveFile.Error())
			}
		}()

		s := FileStorage{filePath: f.Name()}
		require.Error(t, s.LoadData(t.Context(), nil))
	})

	t.Run("correct", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), tmpFilePtrn)
		require.NoError(t, err)
		defer func() {
			if errRemoveFile := os.Remove(f.Name()); errRemoveFile != nil {
				t.Log(errRemoveFile.Error())
			}
		}()

		_, err = f.WriteString(`[
		{"uuid": "1", "short_url": "s1", "original_url": "o1"},
		{"uuid": "2", "short_url": "s2", "original_url": "o2"}
		]`)
		require.NoError(t, err)
		if errClose := f.Close(); errClose != nil {
			t.Log(errClose.Error())
		}
	})
}

func TestSaveData(t *testing.T) {
	const tmpFilePtrn string = "sd*.json"
	f, err := os.CreateTemp(os.TempDir(), tmpFilePtrn)
	require.NoError(t, err)
	defer func() {
		if errRemoveFile := os.Remove(f.Name()); errRemoveFile != nil {
			t.Log(errRemoveFile.Error())
		}
	}()

	stat, err := os.Stat(f.Name())
	require.NoError(t, err)
	require.Equal(t, int64(0), stat.Size())

	fs := FileStorage{filePath: f.Name()}
	resultError := fs.SaveData([]model.StorageRecord{0: {UUID: "1", ShortURL: "s", OrigURL: "o"}})
	require.NoError(t, resultError)

	stat, err = os.Stat(f.Name())
	require.NoError(t, err)
	require.Greater(t, stat.Size(), int64(0))

	fData, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	require.Equal(t, `[{"uuid":"1","short_url":"s","original_url":"o","user_id":"","is_deleted":false}]
`, string(fData))
}

func TestAdd(t *testing.T) {
	patchSave := monkey.PatchInstanceMethod(reflect.TypeOf(&FileStorage{}), "SaveData",
		func(*FileStorage, []model.StorageRecord) error {
			return nil
		})
	defer patchSave.Unpatch()

	patchLoad := monkey.PatchInstanceMethod(reflect.TypeOf(&FileStorage{}), "LoadData",
		func(*FileStorage, context.Context, any) error {
			return nil
		})
	defer patchLoad.Unpatch()

	fs := FileStorage{}
	if err := fs.Add(t.Context(), "1", "s", "o"); err != nil {
		println(err.Error)
	}
}

func ExampleNewStorage() {
	storage, err := NewStorage("data.json")
	fmt.Printf("%+v\t%v", storage, err)

	// Output:
	// &{filePath:data.json}	<nil>
}
