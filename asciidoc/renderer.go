package asciidoc

import (
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// Generate will execute the generation of the documentation
func (p *Producer) Generate() {

	t := NewTemplateWithOverrides(p.overrides)
	w := tabwriter.NewWriter(p.createWriter(), 4, 4, 4, ' ', 0)

	indexdone := !p.index

	err := goparser.ParseSinglePackageWalker(p.parseconfig, func(pkg *goparser.GoPackage) error {

		tc := t.NewContextWithConfig(&pkg.GoFile, &TemplateContextConfig{
			IncludeMethodCode: false,
		})

		if !indexdone {

			ic := tc.DefaultIndexConfig(p.indexconfig)
			if !p.toc {
				ic.TocTitle = "" // disables toc generation
			}

			tc.RenderIndex(w, ic)
			indexdone = true
		}

		tc.RenderPackage(w)

		if len(pkg.Imports) > 0 {
			tc.RenderImports(w)
		}
		if len(pkg.Interfaces) > 0 {
			tc.RenderInterfaces(w)
		}
		if len(pkg.Structs) > 0 {
			tc.RenderStructs(w)
		}
		if len(pkg.CustomTypes) > 0 {
			tc.RenderVarTypeDefs(w)
		}
		if len(pkg.ConstAssignments) > 0 {
			tc.RenderConstDeclarations(w)
		}
		if len(pkg.CustomFuncs) > 0 {
			tc.RenderTypeDefFuncs(w)
		}
		if len(pkg.VarAssigments) > 0 {
			tc.RenderVarDeclarations(w)
		}
		if len(pkg.StructMethods) > 0 {
			tc.RenderFunctions(w)
		}

		return nil

	}, p.paths...)

	if nil != err {
		panic(err)
	}

	w.Flush()
}

type writer struct {
	w io.Writer
}

func (p *Producer) createWriter() io.Writer {

	if p.writer != nil {
		return p.writer
	}

	if p.outfile == "" {
		p.outfile = filepath.Join(p.parseconfig.Module.Base, "docs.adoc")
	}

	dir := filepath.Dir(p.outfile)
	if !dirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}

	wr, err := os.Create(p.outfile)
	if err != nil {
		panic(err)
	}

	return wr

}

func dirExists(dir string) bool {

	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}
