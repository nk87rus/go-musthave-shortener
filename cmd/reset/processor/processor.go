package processor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

func ProcessPackage(dir string) {
	pattern := filepath.Join(dir, "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return
	}

	var pkgName string
	structures := make(map[string]*ast.StructType)
	markedStructs := make(map[string]bool)
	resetMethods := make(map[string]bool)

	for _, file := range files {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			log.Printf("error parsing %s: %v", file, err)
			continue
		}

		// проверка корректности имени пакета
		if pkgName == "" {
			pkgName = node.Name.Name
		} else if pkgName != node.Name.Name {
			log.Printf("warning: package name mismatch in %s: expected %s, got %s", file, pkgName, node.Name.Name)
		}

		// поиск существующих методов Reset()
		collectResetMethods(node, resetMethods)

		// поиск структур с комментарием //generate:reset
		collectResetComments(node, structures, markedStructs)

	}

	if len(markedStructs) == 0 {
		return
	}

	generateResetFile(dir, pkgName, markedStructs, structures, resetMethods)
}

func collectResetMethods(node *ast.File, result map[string]bool) {
	for _, decl := range node.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "Reset" {
			continue
		}
		if fd.Recv == nil || len(fd.Recv.List) == 0 {
			continue
		}
		recv := fd.Recv.List[0].Type
		var typeName string
		switch rt := recv.(type) {
		case *ast.StarExpr:
			if ident, ok := rt.X.(*ast.Ident); ok {
				typeName = ident.Name
			}
		case *ast.Ident:
			typeName = rt.Name
		}
		if typeName != "" {
			result[typeName] = true
		}
	}
}

func collectResetComments(node *ast.File, result map[string]*ast.StructType, markedStructs map[string]bool) {
	for _, decl := range node.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		marked := false
		if gd.Doc != nil {
			for _, c := range gd.Doc.List {
				if strings.Contains(c.Text, "generate:reset") {
					marked = true
					break
				}
			}
		}

		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			specMarked := marked
			if !specMarked && ts.Doc != nil {
				for _, c := range ts.Doc.List {
					if strings.Contains(c.Text, "generate:reset") {
						specMarked = true
						break
					}
				}
			}

			if specMarked {
				markedStructs[ts.Name.Name] = true
			}
			result[ts.Name.Name] = st
		}
	}
}
