package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateShortURLBatch(t *testing.T) {
	var (
		errGRS  = fmt.Errorf("errGetRandomString")
		errRepo = fmt.Errorf("errRepo")
	)
	baseData := []byte(`[
			{"correlation_id": "1", "original_url": "http://one.url"},
			{"correlation_id": "2", "original_url": "http://two.url"}
			]`)
	testCases := []struct {
		name      string
		rawData   []byte
		mFunc     func(m *MockStorage)
		wantError error
	}{
		{
			name:      "ErrDecode",
			rawData:   nil,
			wantError: io.EOF,
		},
		{
			name:      "ErrNoData",
			rawData:   []byte(`[]`),
			wantError: errors.New("отсутсвует значение для обработки"),
		},
		{
			name:      "ErrGetRandomString",
			rawData:   baseData,
			wantError: errGRS,
		},
		{
			name:    "ErrRepoAddBatch",
			rawData: baseData,
			mFunc: func(m *MockStorage) {
				m.On("AddBatch", mock.Anything, mock.Anything).Return(errRepo)
			},
			wantError: errRepo,
		},
		{
			name:    "Correct",
			rawData: baseData,
			mFunc: func(m *MockStorage) {
				m.On("AddBatch", mock.Anything, mock.Anything).Return(nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchGRS := monkey.PatchInstanceMethod(reflect.TypeOf(&Handlers{}), "GetRandomString",
				func(*Handlers, context.Context) (string, error) {
					if errors.Is(tc.wantError, errGRS) {
						return "", tc.wantError
					}
					return "GRS_" + tc.name, nil
				})
			defer patchGRS.Unpatch()

			rMock := NewMockStorage(t)
			if tc.mFunc != nil {
				tc.mFunc(rMock)
			}

			h := &Handlers{repo: rMock, baseURL: &url.URL{}}
			resultData, resultError := h.CreateShortURLBatch(t.Context(), bytes.NewReader(tc.rawData))
			if tc.wantError != nil {
				require.Nil(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.NotNil(t, resultData)
				require.Nil(t, resultError)
			}
		})
	}
}

func TestGetUsersURLs(t *testing.T) {
	var (
		errRepo = fmt.Errorf("errRepo")
	)
	testCases := []struct {
		name      string
		mFunc     func(m *MockStorage)
		wantError error
	}{
		{
			name: "errRepo",
			mFunc: func(m *MockStorage) {
				m.On("GetUsersURLs", mock.Anything).Return(nil, errRepo)
			},
			wantError: errRepo,
		},
		{
			name: "noData",
			mFunc: func(m *MockStorage) {
				m.On("GetUsersURLs", mock.Anything).Return([]model.StorageRecord{}, nil)
			},
			wantError: nil,
		},
		{
			name: "Correct",
			mFunc: func(m *MockStorage) {
				m.On("GetUsersURLs", mock.Anything).
					Return([]model.StorageRecord{{}}, nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rMock := NewMockStorage(t)
			if tc.mFunc != nil {
				tc.mFunc(rMock)
			}

			h := &Handlers{repo: rMock, baseURL: &url.URL{}}
			resultData, resultError := h.GetUsersURLs(t.Context())
			if tc.wantError != nil {
				require.Nil(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				if tc.name != "noData" {
					require.NotNil(t, resultData)
				} else {
					require.Nil(t, resultData)
				}
				require.Nil(t, resultError)
			}
		})
	}
}

func TestDelURLs(t *testing.T) {
	testCases := []struct {
		name      string
		ctx       context.Context
		wantError error
	}{
		{
			name:      "noUID",
			ctx:       t.Context(),
			wantError: nil,
		},
		{
			name:      "Correct",
			ctx:       context.WithValue(t.Context(), model.CtxUserID, "0"),
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rMock := NewMockStorage(t)
			rMock.On("DelURLs", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Maybe()
			h := &Handlers{repo: rMock, baseURL: &url.URL{}}
			h.DelURLs(tc.ctx, nil)
		})
	}
}
