package processor

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAnonymousFieldName(t *testing.T) {
	testCases := []struct {
		name       string
		data       ast.Expr
		wantResult string
	}{
		{
			name:       "Ident",
			data:       &ast.Ident{Name: "IdentName"},
			wantResult: "IdentName",
		},
		{
			name:       "StarExpr",
			data:       &ast.StarExpr{X: &ast.Ident{Name: "StarExpr_Ident"}},
			wantResult: "StarExpr_Ident",
		},
		{
			name:       "SelectorExpr",
			data:       &ast.SelectorExpr{Sel: &ast.Ident{Name: "SelectorExpr_Ident"}},
			wantResult: "SelectorExpr_Ident",
		},
		{
			name:       "DefaultValue",
			data:       &ast.BinaryExpr{},
			wantResult: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantResult, getAnonymousFieldName(tc.data))
		})
	}
}

func TestIsBasic(t *testing.T) {
	testCases := []struct {
		name       string
		value      string
		wantResult bool
	}{
		{
			name:       "IsBasic",
			value:      "int",
			wantResult: true,
		},
		{
			name:       "NotBasic",
			value:      "fake",
			wantResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantResult, isBasicType(tc.value))
		})
	}
}

func TestZeroValue(t *testing.T) {
	testCases := []struct {
		name       string
		value      ast.Expr
		wantResult string
	}{
		{
			name:       "Numeric",
			value:      &ast.Ident{Name: "int"},
			wantResult: "0",
		},
		{
			name:       "Boolean",
			value:      &ast.Ident{Name: "bool"},
			wantResult: "false",
		},
		{
			name:       "String",
			value:      &ast.Ident{Name: "string"},
			wantResult: `""`,
		},
		{
			name:       "Unknown",
			value:      &ast.BadExpr{},
			wantResult: "nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantResult, zeroValue(tc.value))
		})
	}
}
