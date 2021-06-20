// package main contains the one and only binary to run goasciidoc
package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/mariotoffia/goasciidoc/asciidoc"
)

//go:embed defaults/index.gtpl
var templateIndex string

//go:embed defaults/package.gtpl
var templatePackage string

//go:embed defaults/import.gtpl
var templateImports string

//go:embed defaults/functions.gtpl
var templateFunctions string

//go:embed defaults/function.gtpl
var templateFunction string

//go:embed defaults/interface.gtpl
var templateInterface string

//go:embed defaults/interfaces.gtpl
var templateInterfaces string

//go:embed defaults/struct.gtpl
var templateStruct string

//go:embed defaults/structs.gtpl
var templateStructs string

//go:embed defaults/receivers.gtpl
var templateReceivers string

//go:embed defaults/var.gtpl
var templateVarAssignment string

//go:embed defaults/vars.gtpl
var templateVarAssignments string

//go:embed defaults/const.gtpl
var templateConstAssignment string

//go:embed defaults/consts.gtpl
var templateConstAssignments string

//go:embed defaults/typedeffunc.gtpl
var templateCustomFuncDefintion string

//go:embed defaults/typedeffuncs.gtpl
var templateCustomFuncDefintions string

//go:embed defaults/typedefvar.gtpl
var templateCustomTypeDefintion string

//go:embed defaults/typedefvars.gtpl
var templateCustomTypeDefintions string

type args struct {
	Out            string   `arg:"-o"              help:"The out filepath to write the generated document, default module path, file docs.adoc"    placeholder:"PATH"`
	StdOut         bool     `                      help:"If output the generated asciidoc to stdout instead of file"`
	Module         string   `arg:"-m"              help:"an optional folder or file path to module, otherwise current directory"                   placeholder:"PATH"`
	Internal       bool     `arg:"-i"              help:"If internal go code shall be rendered as well"`
	Private        bool     `arg:"-p"              help:"If files beneath directories starting with an underscore shall be included"`
	NonExported    bool     `                      help:"Renders Non exported as well as the exported. Default only Exported is rendered."`
	Test           bool     `arg:"-t"              help:"If test code should be included"`
	NoIndex        bool     `arg:"-n"              help:"If no index header shall be generated"`
	NoToc          bool     `                      help:"Removes the table of contents if index document"`
	IndexConfig    string   `arg:"-c"              help:"JSON document to override the IndexConfig"                                                placeholder:"JSON"`
	Overrides      []string `arg:"-r,separate"     help:"name=template filepath to override default templates"`
	Paths          []string `arg:"positional"      help:"Directory or files to be included in scan (if none, current path is used)"                placeholder:"PATH"`
	ListTemplates  bool     `arg:"--list-template" help:"Lists all default templates in the binary"`
	OutputTemplate string   `arg:"--out-template"  help:"outputs a template to stdout"`
	PackageDoc     []string `arg:"-d,separate"     help:"set relative package search filepaths for package documentation"                          placeholder:"FILEPATH"`
	TemplateDir    string   `                      help:"Loads template files *.gtpl from a directory, use --list to get valid names of templates"`
}

func (args) Version() string {
	return "goasciidoc v0.4.2"
}

func main() {
	var args args
	arg.MustParse(&args)
	runner(args)
}

func runner(args args) {
	if len(args.Paths) == 0 {

		current, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		args.Paths = []string{current}
	}

	p := asciidoc.NewProducer().
		Outfile(args.Out).
		Module(args.Module).
		Include(args.Paths...).
		IndexConfig(args.IndexConfig)

	p.Override(string(asciidoc.ConstDeclarationTemplate), templateConstAssignment)
	p.Override(string(asciidoc.ConstDeclarationsTemplate), templateConstAssignments)
	p.Override(string(asciidoc.FunctionTemplate), templateFunction)
	p.Override(string(asciidoc.FunctionsTemplate), templateFunctions)
	p.Override(string(asciidoc.ImportTemplate), templateImports)
	p.Override(string(asciidoc.IndexTemplate), templateIndex)
	p.Override(string(asciidoc.InterfaceTemplate), templateInterface)
	p.Override(string(asciidoc.InterfacesTemplate), templateInterfaces)
	p.Override(string(asciidoc.PackageTemplate), templatePackage)
	p.Override(string(asciidoc.ReceiversTemplate), templateReceivers)
	p.Override(string(asciidoc.StructTemplate), templateStruct)
	p.Override(string(asciidoc.StructsTemplate), templateStructs)
	p.Override(string(asciidoc.CustomFuncTypeDefTemplate), templateCustomFuncDefintion)
	p.Override(string(asciidoc.CustomFuncTypeDefsTemplate), templateCustomFuncDefintions)
	p.Override(string(asciidoc.CustomVarTypeDefTemplate), templateCustomTypeDefintion)
	p.Override(string(asciidoc.CustomVarTypeDefsTemplate), templateCustomTypeDefintions)
	p.Override(string(asciidoc.VarDeclarationTemplate), templateVarAssignment)
	p.Override(string(asciidoc.VarDeclarationsTemplate), templateVarAssignments)

	p.UseStandardProcessors()

	if args.NoToc {
		p.NoToc()
	}
	if args.Internal {
		p.IncludeInternal()
	}
	if args.Private {
		p.IncludeUnderScoreDirectories()
	}
	if args.Test {
		p.IncludeTest()
	}
	if args.NoIndex {
		p.NoIndex()
	}
	if args.StdOut {
		p.StdOut()
	}

	if len(args.TemplateDir) > 0 {
		files, err := ioutil.ReadDir(args.TemplateDir)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			name := baseName(file.Name())
			p.OverrideFilePath(name, filepath.Join(args.TemplateDir, file.Name()))
		}
	}

	if len(args.Overrides) > 0 {

		for _, o := range args.Overrides {

			kv := strings.Split(o, "=")
			if len(kv) != 2 {
				panic("Overrides must be a name=filepath to template")
			}

			p.OverrideFilePath(kv[0], kv[1])

		}
	}

	if len(args.PackageDoc) > 0 {
		p.PackageDoc(args.PackageDoc...)
	}

	if args.ListTemplates {

		t := p.CreateTemplateWithOverrides()
		for k := range t.Templates {
			fmt.Println(k)
		}

		return
	}

	if args.OutputTemplate != "" {

		if t, ok := p.CreateTemplateWithOverrides().Templates[args.OutputTemplate]; ok {
			fmt.Printf(`"%s"`+"\n", t.Text)
			return
		}

		fmt.Fprintf(os.Stderr, "No template named: %s", args.OutputTemplate)
		return
	}

	if args.NonExported {
		p.NonExported()
	}

	p.Generate()
}

func baseName(s string) string {

	n := strings.LastIndexByte(s, '.')

	if n == -1 {
		return s
	}

	return s[:n]

}
