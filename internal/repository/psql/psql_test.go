package psql

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/jackc/pgx/v5"
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
