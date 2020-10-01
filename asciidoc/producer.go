package asciidoc

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// FileGeneration is which type of file generation to be performed by the Producer
type FileGeneration string

const (
	// GenerateSingle will make the producer to just produce one big file for all packages
	GenerateSingle FileGeneration = "single"
	// GeneratePackage will generate all packages in a individual file (GoPackage)
	GeneratePackage FileGeneration = "package"
	// GenerateFile will genereate one asciidoc file per go file (GoFile)
	GenerateFile FileGeneration = "file"
)

// FileLayout determines which how the generated files are stored, e.g. single directory or
// package structure etc.
type FileLayout string

const (
	// LayoutSingle will store all asciidoc files in a single directory. If duplicate, this will fail
	LayoutSingle FileLayout = "single"
	// LayoutPackage will create (if neccesary) package hiearchies as source files are stored and store
	// the generated in those hiearchies (may be postfixed with additional path see Producer.Postfix()).
	LayoutPackage FileLayout = "package"
)

// Producer parses go code and produces asciidoc documentation.
type Producer struct {
	// parseconfig is the configuration that it uses to invoke
	// the parser with.
	parseconfig goparser.ParseConfig
	// paths is files and directories to include.
	paths []string
	// output is a path where the generated documentation ends up.
	// if output is empty it will emit it directly into source tree.
	output string
	// postfix is one or more folders that gets postfixed onto the
	// package path in order to e.g. have a _docs folder in the source tree
	// for each package.
	postfix string
	// generation stipulates how the files are generated
	generation FileGeneration
	// filelayout determines how the generated files are persistet on the filesystem
	filelayout FileLayout
	// index determines if it will autogenerate a index file that includes all generated documents.
	index bool
}

// NewProducer creates a new instance of a producer.
func NewProducer() *Producer {
	return &Producer{}
}

// Generation specifies how the asciidoc files are generated
func (p *Producer) Generation(gen FileGeneration) *Producer {
	p.generation = gen
	return p
}

// Layout specifies how the generated files are stored onto the filesystem.
func (p *Producer) Layout(layout FileLayout) *Producer {
	p.filelayout = layout
	return p
}

// Index specifies that a asciidoctor index file will be generated
// that includes all generated files.
func (p *Producer) Index() *Producer {
	p.index = true
	return p
}

// Output specifies the output root folder for the documentation.
// If it is set to "" or module root, the documentation will be
// blended into the source code. Use Postfix() to separate it to
// subfolders from package if such is the case.
func (p *Producer) Output(path string) *Producer {
	p.output = path
	return p
}

// Postfix can be used when the resolved path for where to
// write the documentation wishes to be separated. Hence
// it will use the fully qualified path to the package and
// append this postfix.
//
// [TIP]
// .Dont blend source code and documentation
// ====
// If you still want to generate documentation into the source tree,
// use postfix to e.g. set to _docs_ so each package will have _docs_
// folder where the documentation is rendered.
// ====
func (p *Producer) Postfix(postfix string) *Producer {
	p.postfix = postfix
	return p
}

// Module directs the producer to pick up module from path.
//
// path may be a directory or a full path to go.mod. If "" it
// will use current directory.
func (p *Producer) Module(path string) *Producer {

	if "" == path {

		d, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		path = d
	}

	if !strings.HasSuffix(path, "go.mod") {
		path = filepath.Join(path, "go.mod")
	}

	m, err := goparser.NewModule(path)
	if err != nil {
		panic(err)
	}

	p.parseconfig.Module = m

	if p.output == "" {
		p.output = p.parseconfig.Module.Base
	}

	return p
}

// Include adds one or more directory or files in any combination. The producer
// will sort out which are directories and which are filepaths.
//
// If filepath, it will not do any type of checking and will blindly think it is a
// valid go file.
func (p *Producer) Include(path ...string) *Producer {
	p.paths = append(p.paths, path...)
	return p
}

// IncludeTest will create documentation for test files as well.
func (p *Producer) IncludeTest() *Producer {
	p.parseconfig.Test = true
	return p
}

// IncludeInternal will include internal folder source files.
func (p *Producer) IncludeInternal() *Producer {
	p.parseconfig.Internal = true
	return p
}

// IncludeUnderScoreDirectories will include files that resides below
// directories starting with underscore.
func (p *Producer) IncludeUnderScoreDirectories() *Producer {
	p.parseconfig.UnderScore = true
	return p
}

// target renders a out directory path where the documentation may be written.
func (p *Producer) target(pkg *goparser.GoPackage) string {
	relpkg := pkg.Path[len(pkg.Module.Base):]
	outpath := p.output

	if p.postfix != "" {
		return filepath.Join(outpath, relpkg, p.postfix)
	}

	return filepath.Join(outpath, relpkg)
}
