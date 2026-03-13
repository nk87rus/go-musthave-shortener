// Package noexit – собственный анализатор, запрещающий прямой вызов os.Exit
// в функции main пакета main.
//
package noexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer - запрещает вызов os.Exit в функции main пакета main
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "disallow os.Exit call in main function of main package",
	Run:  runNoOsExit,
}

// runNoOsExit реализует логику анализатора noexit.
// Он обходит AST каждого файла в пакете main, находит функцию main и проверяет,
// содержит ли её тело прямой вызов os.Exit. Если такой вызов найден,
// анализатор сообщает о диагностике.
func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// Работаем только с пакетом main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Поиск функции main
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			// Анализ тела функции main
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if isOsExitCall(call, pass.TypesInfo) {
					pass.Reportf(call.Pos(), "direct call to os.Exit from main function is forbidden")
				}
				return true
			})
			return false // останавливаем обход после нахождения main
		})
	}
	return nil, nil
}

// isOsExitCall проверяет, является ли выражение вызовом os.Exit.
// Используется информация о типах, чтобы отличить настоящий пакет os от других
// идентификаторов с именем "os".
func isOsExitCall(call *ast.CallExpr, info *types.Info) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Проверка имени пакета и функции
	if pkgIdent, ok := sel.X.(*ast.Ident); ok {
		if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
			// Убеждаемся, что это действительно пакет "os"
			if obj := info.ObjectOf(pkgIdent); obj != nil {
				if pkgName, ok := obj.(*types.PkgName); ok {
					return pkgName.Imported().Path() == "os"
				}
			}
		}
	}
	return false
}
