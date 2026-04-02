package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nk87rus/go-musthave-shortener/cmd/reset/processor"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if isIgnoredDir(d.Name()) {
			return filepath.SkipDir
		}

		files, err := filepath.Glob(filepath.Join(path, "*.go"))
		if err != nil || len(files) == 0 {
			return nil
		}

		processor.ProcessPackage(path)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func isIgnoredDir(name string) bool {
	return name == "vendor" || (strings.HasPrefix(name, ".") && len(name) > 1)
}
