package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/jackc/pgx/v5"
	"github.com/nk87rus/go-musthave-shortener/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	var errApplyMigrations = fmt.Errorf("errAM")
	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errApplyMigrations",
			wantError: errApplyMigrations,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dMock := NewMockPSQLDriver(t)
			dMock.On("GetConnConfig").Return(nil).Maybe()

			patchAM := monkey.Patch(applyMigrations,
				func(context.Context, *pgx.ConnConfig) error {
					if errors.Is(tc.wantError, errApplyMigrations) {
						return tc.wantError
					}
					return nil
				})
			defer patchAM.Unpatch()

			resultData, resultError := NewStorage(t.Context(), dMock)
			if tc.wantError != nil {
				require.Nil(t, resultData)
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
				require.NotEmpty(t, resultData)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	dMock := NewMockPSQLDriver(t)
	dMock.On("Insert", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	resultError := (&PSQLStorage{db: dMock}).Add(context.WithValue(t.Context(), model.CtxUserID, "test"), "1", "s", "o")
	require.Nil(t, resultError)
}

func TestLoadData(t *testing.T) {
	var (
		errSelectBytes = fmt.Errorf("errSB")
		errUM          = fmt.Errorf("errUM")
	)

	testCases := []struct {
		name      string
		mFunc     func(m *MockPSQLDriver)
		wantError error
	}{
		{
			name: "errSelectBytes",
			mFunc: func(m *MockPSQLDriver) {
				m.On("SelectBytes", mock.Anything, mock.AnythingOfType("string")).Return(nil, errSelectBytes)
			},
			wantError: errSelectBytes,
		},
		{
			name: "errUM",
			mFunc: func(m *MockPSQLDriver) {
				m.On("SelectBytes", mock.Anything, mock.AnythingOfType("string")).Return([]byte{}, nil)
			},
			wantError: errUM,
		},
		{
			name: "Correct",
			mFunc: func(m *MockPSQLDriver) {
				m.On("SelectBytes", mock.Anything, mock.AnythingOfType("string")).Return([]byte{}, nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchJUM := monkey.Patch(json.Unmarshal,
				func(d []byte, v any) error {
					if errors.Is(tc.wantError, errUM) {
						return tc.wantError
					}
					return nil
				})
			defer patchJUM.Unpatch()

			dMock := NewMockPSQLDriver(t)
			if tc.mFunc != nil {
				tc.mFunc(dMock)
			}

			resultError := (&PSQLStorage{db: dMock}).LoadData(t.Context(), nil)
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
			}
		})
	}
}
