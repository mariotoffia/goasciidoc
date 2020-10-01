package asciidoc

import (
	"io"
	"os"
	"path/filepath"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// Generate will execute the generation of the documentation
func (p *Producer) Generate() {

	t := NewTemplateWithOverrides(p.overrides)
	w := p.createWriter()

	indexdone := !p.index

	err := goparser.ParseSinglePackageWalker(p.parseconfig, func(pkg *goparser.GoPackage) error {

		tc := t.NewContextWithConfig(&pkg.GoFile, &TemplateContextConfig{
			IncludeMethodCode: false,
		})

		if !indexdone {
			tc.RenderIndex(w, tc.DefaultIndexConfig(p.indexconfig))
			indexdone = true
		}

		tc.RenderPackage(w)
		tc.RenderImports(w)
		tc.RenderInterfaces(w)
		tc.RenderStructs(w)
		tc.RenderVarTypeDefs(w)
		tc.RenderConstDeclarations(w)
		tc.RenderTypeDefFuncs(w)
		tc.RenderVarDeclarations(w)
		tc.RenderFunctions(w)

		return nil

	}, p.paths...)

	if nil != err {
		panic(err)
	}
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
