package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	var (
		errUserID      = fmt.Errorf("не корректный тип userID")
		errGetShortURL = fmt.Errorf("errGetShortURL")
		errInsert1     = fmt.Errorf("errInsert1")
		errInsert2     = fmt.Errorf("errInsert2")
	)
	testCases := []struct {
		name      string
		ctx       context.Context
		mFunc     func(m *MockPSQLDriver)
		wantError error
	}{
		{
			name:      "wrongUserID",
			ctx:       context.WithValue(t.Context(), model.CtxUserID, 1),
			wantError: errUserID,
		},
		{
			name: "pgErr.UniqueViolation_1",
			ctx:  context.WithValue(t.Context(), model.CtxUserID, "test"),
			mFunc: func(m *MockPSQLDriver) {
				m.On("Insert",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string")).
					Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				m.On("SelectString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("", errGetShortURL)

			},
			wantError: errGetShortURL,
		},
		{
			name: "pgErr.UniqueViolation_2",
			ctx:  context.WithValue(t.Context(), model.CtxUserID, "test"),
			mFunc: func(m *MockPSQLDriver) {
				m.On("Insert",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string")).
					Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: errInsert1.Error()})
				m.On("SelectString", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("test", nil)

			},
			wantError: errInsert1,
		},
		{
			name: "errInsert2",
			ctx:  context.WithValue(t.Context(), model.CtxUserID, "test"),
			mFunc: func(m *MockPSQLDriver) {
				m.On("Insert",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string")).
					Return(errInsert2)

			},
			wantError: errInsert2,
		},
		{
			name: "Correct",
			ctx:  context.WithValue(t.Context(), model.CtxUserID, "test"),
			mFunc: func(m *MockPSQLDriver) {
				m.On("Insert",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string")).
					Return(nil)

			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dMock := NewMockPSQLDriver(t)
			if tc.mFunc != nil {
				tc.mFunc(dMock)
			}

			resultError := (&PSQLStorage{db: dMock}).Add(tc.ctx, "1", "s", "o")

			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
			}
		})
	}
}

func TestAddBatch(t *testing.T) {
	var errInsert = fmt.Errorf("errInsert")
	testCases := []struct {
		name      string
		mFunc     func(m *MockPSQLDriver)
		wantError error
	}{
		{
			name: "errInsert",
			mFunc: func(m *MockPSQLDriver) {
				m.On("InsertBatch", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(errInsert)
			},
			wantError: errInsert,
		},
		{
			name: "Correct",
			mFunc: func(m *MockPSQLDriver) {
				m.On("InsertBatch", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(errInsert)
			},
			wantError: errInsert,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dMock := NewMockPSQLDriver(t)
			if tc.mFunc != nil {
				tc.mFunc(dMock)
			}

			resultError := (&PSQLStorage{db: dMock}).AddBatch(t.Context(), slices.Values([]model.StorageRecord{{}}))

			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
			}
		})
	}
}

func TestReqTimeout(t *testing.T) {
	testCases := []struct {
		name       string
		data       int
		wantResult time.Duration
	}{
		{
			name:       "Less_5s",
			data:       1,
			wantResult: 5 * time.Second,
		},
		{
			name:       "Gt_5s",
			data:       8,
			wantResult: 12 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantResult, reqTimeout(tc.data))
		})
	}
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

func TestDelURLs(t *testing.T) {
	dMock := NewMockPSQLDriver(t)
	dMock.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	require.Nil(t, (&PSQLStorage{db: dMock}).DelURLs(t.Context(), "", nil))
}
