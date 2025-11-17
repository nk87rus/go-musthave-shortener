package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestInitPSQL(t *testing.T) {
	var errConn = fmt.Errorf("errConnect")
	testCases := []struct {
		name      string
		wantError error
	}{
		{
			name:      "errConn",
			wantError: errConn,
		},
		{
			name:      "Correct",
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patchConnect := monkey.Patch(pgx.Connect,
				func(context.Context, string) (*pgx.Conn, error) {
					if errors.Is(tc.wantError, errConn) {
						return nil, tc.wantError
					}
					return new(pgx.Conn), nil
				})
			defer patchConnect.Unpatch()

			resultData, resultError := InitPSQL(context.Background(), "")
			if tc.wantError != nil {
				require.ErrorContains(t, resultError, tc.wantError.Error())
				require.Nil(t, resultData)
			} else {
				require.Nil(t, resultError)
				require.IsType(t, &PSQL{}, resultData)
			}
		})
	}
}

func TestPSQLClose(t *testing.T) {
	patchClose := monkey.PatchInstanceMethod(reflect.TypeOf(&pgx.Conn{}), "Close",
		func(*pgx.Conn, context.Context) error {
			return fmt.Errorf("test")
		})
	defer patchClose.Unpatch()
	require.Error(t, (&PSQL{}).Close(context.Background()))
}

// func TestPing(t *testing.T) {
// 	t.Run("pgErr", func(t *testing.T) {
// 		patchPing := monkey.PatchInstanceMethod(reflect.TypeOf(&pgx.Conn{}), "Ping",
// 			func(*pgx.Conn, context.Context) error {
// 				return &pgconn.PgError{Message: "test pgErr"}
// 			})
// 		defer patchPing.Unpatch()

// 		err := (&PSQL{conn: &pgx.Conn{}}).Ping(context.Background())
// 		require.Equal(t, err.Error(), "test pgErr")
// 	})

// 	t.Run("basicErr", func(t *testing.T) {
// 		patchPing := monkey.PatchInstanceMethod(reflect.TypeOf(&pgx.Conn{}), "Ping",
// 			func(*pgx.Conn, context.Context) error {
// 				return fmt.Errorf("bErr")
// 			})
// 		defer patchPing.Unpatch()

// 		err := (&PSQL{conn: &pgx.Conn{}}).Ping(context.Background())
// 		require.Equal(t, err.Error(), "bErr")
// 	})
// }
