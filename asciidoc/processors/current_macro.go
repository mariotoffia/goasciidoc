package processors

import (
	"path/filepath"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// CurrentMacroProcessor handles _${gad:current:macro_name}_ macros.
func CurrentMacroProcessor() DocProcessor {

	return func(module *goparser.GoModule, file *goparser.GoFile, doc, fq string) string {

		fqdir := filepath.Dir(fq)
		dir := filepath.Base(fqdir)
		basefile := filepath.Base(fq)

		doc = strings.ReplaceAll(doc, "${gad:current:fq}", fq)
		doc = strings.ReplaceAll(doc, "${gad:current:fqdir}", fqdir)
		doc = strings.ReplaceAll(doc, "${gad:current:dir}", dir)
		return strings.ReplaceAll(doc, "${gad:current:file}", basefile)
	}

}
