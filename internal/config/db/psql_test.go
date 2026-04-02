//go:generate go run github.com/vektra/mockery/v2 --name=Tx --output=./ --outpkg=db --filename=mock_Tx_test.go --dir=$GOMODCACHE/github.com/jackc/pgx/v5@v5.7.0
//go:generate go run github.com/vektra/mockery/v2 --name=BatchResults --output=./ --outpkg=db --filename=mock_BatchResults_test.go --dir=$GOMODCACHE/github.com/jackc/pgx/v5@v5.7.0
//go:generate go run github.com/vektra/mockery/v2 --name=Row --output=./ --outpkg=db --filename=mock_Row_test.go --dir=$GOMODCACHE/github.com/jackc/pgx/v5@v5.7.0
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// func TestInitPSQL(t *testing.T) {
// 	var errConn = fmt.Errorf("errConnect")
// 	testCases := []struct {
// 		name      string
// 		wantError error
// 	}{
// 		{
// 			name:      "errConn",
// 			wantError: errConn,
// 		},
// 		{
// 			name:      "Correct",
// 			wantError: nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			patchConnect := monkey.Patch(pgx.Connect,
// 				func(context.Context, string) (*pgx.Conn, error) {
// 					if errors.Is(tc.wantError, errConn) {
// 						return nil, tc.wantError
// 					}
// 					return new(pgx.Conn), nil
// 				})
// 			defer patchConnect.Unpatch()

// 			resultData, resultError := InitPSQL(t.Context(), "")
// 			if tc.wantError != nil {
// 				require.ErrorContains(t, resultError, tc.wantError.Error())
// 				require.Nil(t, resultData)
// 			} else {
// 				require.Nil(t, resultError)
// 				require.IsType(t, &PSQL{}, resultData)
// 			}
// 		})
// 	}
// }

func TestGracefulShutdown(t *testing.T) {
	testCases := []struct {
		name  string
		mFunc func(m *MockDBConn)
	}{
		{
			name: "errOnClose",
			mFunc: func(m *MockDBConn) {
				m.On("IsClosed").Return(false)
				m.On("Close", mock.Anything).Return(fmt.Errorf("errOnClose"))
			},
		},
		{
			name: "Correct",
			mFunc: func(m *MockDBConn) {
				m.On("IsClosed").Return(true)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cMock := NewMockDBConn(t)
			tc.mFunc(cMock)

			ctx, fCancel := context.WithCancel(t.Context())
			fCancel()

			p := PSQL{conn: cMock}
			p.GracefulShutdown(ctx)
		})
	}
}

func TestPSQLClose(t *testing.T) {
	cMock := NewMockDBConn(t)
	cMock.On("IsClosed").Return(false)
	cMock.On("Close", mock.Anything).Return(fmt.Errorf("OK"))
	p := &PSQL{conn: cMock}
	require.Error(t, p.Close(context.Background()))
}

func TestPing(t *testing.T) {
	var (
		pgErr    = &pgconn.PgError{Message: "test pgErr"}
		basicErr = fmt.Errorf("bErr")
	)
	testCases := []struct {
		name      string
		mFunc     func(m *MockDBConn)
		wantError error
	}{
		{
			name: "pgErr",
			mFunc: func(m *MockDBConn) {
				m.On("Ping", mock.Anything).Return(pgErr)
			},
			wantError: pgErr,
		},
		{
			name: "basicErr",
			mFunc: func(m *MockDBConn) {
				m.On("Ping", mock.Anything).Return(basicErr)
			},
			wantError: basicErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cMock := NewMockDBConn(t)
			tc.mFunc(cMock)

			p := PSQL{conn: cMock}
			require.ErrorContains(t, tc.wantError, p.Ping(t.Context()).Error())
		})
	}
}

func TestGetConnConfig(t *testing.T) {
	cMock := NewMockDBConn(t)
	cMock.On("Config").Return(&pgx.ConnConfig{DescriptionCacheCapacity: 1})
	p := PSQL{conn: cMock}
	require.NotEmpty(t, p.GetConnConfig())
}

func TestInsert(t *testing.T) {
	var errExec = fmt.Errorf("errExec")
	testCases := []struct {
		name      string
		mFunc     func(m *MockDBConn)
		wantError error
	}{
		{
			name: "errExec",
			mFunc: func(m *MockDBConn) {
				m.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(pgconn.CommandTag{}, errExec)
			},
			wantError: errExec,
		},
		{
			name: "Correct",
			mFunc: func(m *MockDBConn) {
				m.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(pgconn.CommandTag{}, nil)
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cMock := NewMockDBConn(t)
			tc.mFunc(cMock)
			p := PSQL{conn: cMock}

			resultError := p.Insert(t.Context(), "")
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
			} else {
				require.Nil(t, resultError)
			}
		})
	}
}

func TestInsertBatch(t *testing.T) {
	var (
		errBegin     = fmt.Errorf("errBegin")
		errSendBatch = fmt.Errorf("errSendBatch")
	)
	testCases := []struct {
		name      string
		cmFunc    func(cm *MockDBConn)
		wantError error
	}{
		{
			name: "errBegin",
			cmFunc: func(cm *MockDBConn) {
				cm.On("Begin", mock.Anything).Return(nil, errBegin)
			},
			wantError: errBegin,
		},
		{
			name: "errSendBatch",
			cmFunc: func(cm *MockDBConn) {
				brMock := NewBatchResults(t)
				brMock.On("Close").Return(fmt.Errorf("errCloseBatchResults"))
				brMock.On("Exec").Return(pgconn.CommandTag{}, errSendBatch)

				txMock := NewTx(t)
				txMock.On("SendBatch", mock.Anything, mock.Anything).Return(brMock)
				txMock.On("Rollback", mock.Anything).Return(nil)
				cm.On("Begin", mock.Anything).Return(txMock, nil)
			},
			wantError: errSendBatch,
		},
		{
			name: "Correct",
			cmFunc: func(cm *MockDBConn) {
				brMock := NewBatchResults(t)
				brMock.On("Close").Return(nil)
				brMock.On("Exec").Return(pgconn.CommandTag{}, nil)

				txMock := NewTx(t)
				txMock.On("SendBatch", mock.Anything, mock.Anything).Return(brMock)
				txMock.On("Commit", mock.Anything).Return(nil)
				cm.On("Begin", mock.Anything).Return(txMock, nil)
			},
			wantError: nil,
		},
	}

	na := make([]pgx.NamedArgs, 0, 1001)
	for i := 0; i < cap(na); i++ {
		na = append(na, pgx.NamedArgs{"foo": i})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmMock := NewMockDBConn(t)
			if tc.cmFunc != nil {
				tc.cmFunc(cmMock)
			}

			p := PSQL{conn: cmMock}
			resultError := p.InsertBatch(t.Context(), "", na)
			if tc.wantError != nil {
				require.Error(t, resultError)
			} else {
				require.Nil(t, resultError)
			}
		})
	}
}

func TestSelectBytes(t *testing.T) {
	cMock := NewMockDBConn(t)
	cMock.On("QueryRow", mock.Anything, mock.Anything).Return(
		func() pgx.Row {
			rMock := NewRow(t)
			rMock.On("Scan", mock.Anything).
				Return(nil).
				Run(func(args mock.Arguments) {
					a := args.Get(0).(*[]byte)
					*a = []byte{1}
				})
			return rMock
		}(),
	)

	p := PSQL{conn: cMock}
	resultData, resultError := p.SelectBytes(t.Context(), "")
	require.NotEmpty(t, resultData)
	require.Nil(t, resultError)
}

func TestSelectString(t *testing.T) {
	cMock := NewMockDBConn(t)
	cMock.On("QueryRow", mock.Anything, mock.Anything).Return(
		func() pgx.Row {
			rMock := NewRow(t)
			rMock.On("Scan", mock.Anything).
				Return(nil).
				Run(func(args mock.Arguments) {
					a := args.Get(0).(*string)
					*a = "test"
				})
			return rMock
		}(),
	)

	p := PSQL{conn: cMock}
	resultData, resultError := p.SelectString(t.Context(), "")
	require.NotEmpty(t, resultData)
	require.Nil(t, resultError)
}

func TestExec(t *testing.T) {
	cMock := NewMockDBConn(t)
	cMock.On("Exec", mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)
	p := PSQL{conn: cMock}
	require.Nil(t, p.Exec(t.Context(), ""))

}
