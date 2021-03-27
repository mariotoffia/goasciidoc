package asciidoc

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// getProcessMacroFunc returns the macro substitution function.
func getProcessMacroFunc(
	next goparser.ParseSinglePackageWalkerFunc) goparser.ParseSinglePackageWalkerFunc {

	return func(pkg *goparser.GoPackage) error {

		// Clean out asciidoc tags from the source
		r, _ := regexp.Compile(`\s*tag::.*\[\]`)

		processDocs := func(doc, fq string) string {

			fqdir := filepath.Dir(fq)
			dir := filepath.Base(fqdir)
			file := filepath.Base(fq)

			doc = strings.ReplaceAll(doc, "${gad:current:fq}", fq)
			doc = strings.ReplaceAll(doc, "${gad:current:fqdir}", fqdir)
			doc = strings.ReplaceAll(doc, "${gad:current:dir}", dir)
			doc = strings.ReplaceAll(doc, "${gad:current:file}", file)

			return r.ReplaceAllString(doc, "")
		}

		pkg.Doc = processDocs(pkg.Doc, pkg.Module.Base)

		for _, c := range pkg.ConstAssignments {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.CustomFuncs {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.CustomTypes {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Imports {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Interfaces {
			c.Doc = processDocs(c.Doc, c.File.FilePath)

			for _, c := range c.Methods {
				c.Doc = processDocs(c.Doc, c.File.FilePath)
			}

		}

		for _, c := range pkg.StructMethods {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Structs {
			c.Doc = processDocs(c.Doc, c.File.FilePath)

			for _, c := range c.Fields {
				c.Doc = processDocs(c.Doc, c.File.FilePath)
			}

		}

		for _, c := range pkg.VarAssignments {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		return next(pkg)
	}

}
