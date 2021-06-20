package processors

import (
	"github.com/mariotoffia/goasciidoc/goparser"
)

// DocProcessor is a processor that may alter the documentation by processing and returning the
// *complete* replacement of the documentation.
//
// NOTE: The _file_ parameter is `nil` when a package block is processed. The _fq_ is then not a
// file as well, instead the directory to the package.
type DocProcessor func(module *goparser.GoModule, file *goparser.GoFile, doc, fq string) string

// DocProcessorRegistry is a registry that one may register one or more
// `DocProcessor` functions.
//
// It is then able to do the actual processing via `GetDocProcessorFunc` function.
type DocProcessorRegistry struct {
	processors []DocProcessor
}

func NewDocProcessorRegistry() *DocProcessorRegistry {

	return &DocProcessorRegistry{
		processors: []DocProcessor{},
	}

}

func (p *DocProcessorRegistry) IsEmpty() bool {

	return len(p.processors) == 0

}

func (p *DocProcessorRegistry) Register(proc ...DocProcessor) *DocProcessorRegistry {

	p.processors = append(p.processors, proc...)
	return p

}

// getProcessMacroFunc returns the macro substitution function.
func (p *DocProcessorRegistry) GetProcessMacroFunc(
	next goparser.ParseSinglePackageWalkerFunc) goparser.ParseSinglePackageWalkerFunc {

	return func(pkg *goparser.GoPackage) error {

		processDocs := func(module *goparser.GoModule, file *goparser.GoFile, doc, fq string) string {

			for _, p := range p.processors {

				doc = p(module, file, doc, fq)

			}

			return doc

		}

		pkg.Doc = processDocs(pkg.Module, nil, pkg.Doc, pkg.Module.Base)

		for _, c := range pkg.ConstAssignments {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.CustomFuncs {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.CustomTypes {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Imports {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Interfaces {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)

			for _, c := range c.Methods {
				c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
			}

		}

		for _, c := range pkg.StructMethods {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Structs {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)

			for _, c := range c.Fields {
				c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
			}

		}

		for _, c := range pkg.VarAssignments {
			c.Doc = processDocs(c.File.Module, c.File, c.Doc, c.File.FilePath)
		}

		return next(pkg)
	}

}
