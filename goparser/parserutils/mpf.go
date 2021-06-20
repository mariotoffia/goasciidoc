package parserutils

import "fmt"

// ModulePathAndFiles is keyed with module path and `PathAndFiles` for
// that module.
type ModulePathAndFiles map[string]PathAndFiles

func FromPathAndFiles(pf PathAndFiles) ModulePathAndFiles {

	mpf := ModulePathAndFiles{}

	mods := pf.ModulePaths()
	for _, mod := range mods {

		mpf[mod] = PathAndFiles{}

	}

	for fp, files := range pf {
		fmt.Println(fp)
		fmt.Println(files)

	}

	return mpf
}
