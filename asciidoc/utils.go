package asciidoc

import (
	"os"
	"path/filepath"
)

func dirExists(dir string) bool {

	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

func fileExists(filepath string) bool {

	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func currentDir(element string) string {

	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(d, element)
}
