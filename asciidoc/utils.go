package asciidoc

import "os"

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
