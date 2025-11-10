package asciidoc

import (
	"fmt"
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
	// overviewpaths is which paths to search for overview ascii doc document.
	// It defaults to overview.adoc, _design/overview.adoc.
	overviewpaths []string
	// private when set to true all symbols are rendered.
	private bool
	// macro determine if a additional pass is done to substitute ${goasciidoc:macroname:...} with
	// corresponding values.
	macro bool
	// typeLinks controls linking behaviour for referenced types.
	typeLinks TypeLinkMode
	// signatureStyle controls how signatures are rendered.
	signatureStyle string
	// highlighter controls which source highlighter attribute is emitted in the index header.
	highlighter string
	// renderOptions controls what examples to render (struct-json, struct-yaml).
	renderOptions map[string]bool
}

// NewProducer creates a new instance of a producer.
func NewProducer() *Producer {
	return &Producer{
		overrides:      map[string]string{},
		index:          true,
		typeLinks:      TypeLinksDisabled,
		signatureStyle: "source",
	}
}

// Debug toggles debug logging to stdout.
func (p *Producer) Debug(enabled bool) *Producer {
	if !enabled {
		p.parseconfig.Debug = nil
		return p
	}

	p.parseconfig.Debug = func(format string, args ...interface{}) {
		fmt.Fprintf(os.Stdout, "[debug] "+format+"\n", args...)
	}

	return p
}

func (p *Producer) debugf(format string, args ...interface{}) {
	if p.parseconfig.Debug == nil {
		return
	}

	p.parseconfig.Debug("asciidoc: "+format, args...)
}

// StdOut writes to stdout instead onto filesystem.
func (p *Producer) StdOut() *Producer {
	p.writer = os.Stdout
	return p
}

// EnableMacro will enable the substitution of _goasciidoc_ custom macros.
//
// A _goasciidoc_ macro is on the following form _${gad:macro-name[:...]}_ footnote:[goad stands for goasciidoc].
func (p *Producer) EnableMacro() *Producer {
	p.macro = true
	return p
}

// NonExported will set renderer to render all Symbols both
// exported and non exported. By default only exported symbols
// are rendered.
func (p *Producer) NonExported() *Producer {
	p.private = true
	return p
}

// Writer sets a custom writer where *everything* gets written to.
func (p *Producer) Writer(w io.Writer) *Producer {
	p.writer = w
	return p
}

// PackageDoc adds a relative, each package, filepath to search for overview package asciidoc.
//
// For example _design/package.adoc will make goasciidoc to search relative each package path
// for this particular folder and file.
func (p *Producer) PackageDoc(filepath ...string) *Producer {
	p.overviewpaths = append(p.overviewpaths, filepath...)
	return p
}

// OverrideFilePath will use another template instead of a built-in default
// for the particular name (see TemplateType for valid template names)
// This is loaded from the in param path.
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

// NoIndex specifies that the generated asciidoctor document will not have
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
// will override the default IndexConfig configuration. If no override,
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

	if path == "" {

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

// TypeLinks configures how type references are rendered inside the generated documentation.
func (p *Producer) TypeLinks(mode TypeLinkMode) *Producer {
	p.typeLinks = mode
	return p
}

// Concatenation configures how doc comments split by blank lines are combined.
func (p *Producer) Concatenation(mode goparser.DocConcatenationMode) *Producer {
	p.parseconfig.DocConcatenation = mode
	return p
}

// SignatureStyle controls how signatures are rendered (e.g. "goasciidoc", "source").
func (p *Producer) SignatureStyle(style string) *Producer {
	p.signatureStyle = strings.TrimSpace(strings.ToLower(style))
	return p
}

// Highlighter controls which source highlighter attribute is emitted in the index header.
func (p *Producer) Highlighter(name string) *Producer {
	p.highlighter = strings.TrimSpace(name)
	return p
}

// RenderOptions controls what examples to render for structs (struct-json, struct-yaml).
func (p *Producer) RenderOptions(opts map[string]bool) *Producer {
	p.renderOptions = opts
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

// BuildTags sets specific build tags to use when loading packages.
// Each tag can be a comma-separated list (e.g., "linux,amd64").
func (p *Producer) BuildTags(tags ...string) *Producer {
	p.parseconfig.BuildTags = tags
	return p
}

// AllBuildTags enables auto-discovery of all build tags in the source code.
func (p *Producer) AllBuildTags(enabled bool) *Producer {
	p.parseconfig.AllBuildTags = enabled
	return p
}

// IgnoreMarkdownHeadings enables replacement of markdown headings (#, ##, etc.) with their text content.
func (p *Producer) IgnoreMarkdownHeadings(enabled bool) *Producer {
	p.parseconfig.IgnoreMarkdownHeadings = enabled
	return p
}
