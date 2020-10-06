package asciidoc

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// Producer parses go code and produces asciidoc documentation.
type Producer struct {
	// parseconfig is the configuration that it uses to invoke
	// the parser with.
	parseconfig goparser.ParseConfig
	// paths is files and directories to include.
	paths []string
	// outfile is the file to write the generated documentation onto
	outfile string
	// index determines if it will render index as header for all
	// rendered documents. If inclusion, this might be a good idea
	// not to render index. Default is true.
	index bool
	// indexconfig is a JSON document to override the default IndexConfig
	// when rendering the index template
	indexconfig string
	// overrides is the template overrides that is passed to the template engine.
	overrides map[string]string
	// writer is a fixed custom writer that *all* gets written to.
	writer io.Writer
	// toc enables or disables the table of contents if index is set to true
	// default is true
	toc bool
}

// NewProducer creates a new instance of a producer.
func NewProducer() *Producer {
	return &Producer{
		overrides: map[string]string{},
		index:     true,
	}
}

// StdOut writes to stdout instead onto filesystem.
func (p *Producer) StdOut() *Producer {
	p.writer = os.Stdout
	return p
}

// Writer sets a custom writer where *everything* gets written to.
func (p *Producer) Writer(w io.Writer) *Producer {
	p.writer = w
	return p
}

// OverrideFilePath will use another template instead of a built-in default
// for the particular name (see TemplateType for valid template names)
// This is loaded from the inparam path.
func (p *Producer) OverrideFilePath(name, path string) *Producer {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return p.Override(name, string(data))
}

// Override will use another template instead of a built-in default
// for the particular name (see TemplateType for valid template names)
func (p *Producer) Override(name, template string) *Producer {
	p.overrides[name] = template
	return p
}

// Outfile sets a file to write to
func (p *Producer) Outfile(path string) *Producer {
	p.outfile = path
	return p
}

// NoIndex specifies that the genereated asciidoctor document will not have
// a index header. This is good for inclusion where a header is already present.
func (p *Producer) NoIndex() *Producer {
	p.index = false
	return p
}

// NoToc disables the table of contents if index is enabled. Default
// is when index is enabled a table of contents is produced.
func (p *Producer) NoToc() *Producer {
	p.toc = false
	return p
}

// IndexConfig will configures using SON properties and hence it
// will override the default IndexConfig configuration. If no overide,
// just pass an empty string.
func (p *Producer) IndexConfig(overrides string) *Producer {
	p.indexconfig = overrides
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

		path = filepath.Join(d, "go.mod")
	}

	if !strings.HasSuffix(path, "go.mod") {
		path = filepath.Join(path, "go.mod")
	}

	m, err := goparser.NewModule(path)
	if err != nil {
		panic(err)
	}

	p.parseconfig.Module = m

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
