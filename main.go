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
	"github.com/mariotoffia/goasciidoc/goparser"
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
var templateCustomFuncDefinitions string

//go:embed defaults/typedefvar.gtpl
var templateCustomTypeDefintion string

//go:embed defaults/typedefvars.gtpl
var templateCustomTypeDefinitions string

//go:embed defaults/module.gtpl
var templateModule string

type args struct {
	Out                    string   `arg:"-o"                         help:"The out filepath to write the generated document, default module path, file docs.adoc"                    placeholder:"PATH"`
	StdOut                 bool     `                                 help:"If output the generated asciidoc to stdout instead of file"`
	Debug                  bool     `arg:"--debug"                    help:"Outputs debug statements to stdout during processing"`
	Module                 string   `arg:"-m"                         help:"an optional folder or file path to module, otherwise current directory"                                   placeholder:"PATH"`
	Internal               bool     `arg:"-i"                         help:"If internal go code shall be rendered as well"`
	Private                bool     `arg:"-p"                         help:"If files beneath directories starting with an underscore shall be included"`
	NonExported            bool     `                                 help:"Renders Non exported as well as the exported. Default only Exported is rendered."`
	Test                   bool     `arg:"-t"                         help:"If test code should be included"`
	NoIndex                bool     `arg:"-n"                         help:"If no index header shall be generated"`
	NoToc                  bool     `                                 help:"Removes the table of contents if index document"`
	IndexConfig            string   `arg:"-c"                         help:"JSON document to override the IndexConfig"                                                                placeholder:"JSON"`
	Overrides              []string `arg:"-r,separate"                help:"name=template filepath to override default templates"`
	Paths                  []string `arg:"positional"                 help:"Directory or files to be included in scan (if none, current path is used)"                                placeholder:"PATH"`
	Excludes               []string `arg:"--exclude,separate"         help:"Regex or glb: prefixed glob-like patterns to exclude paths (e.g., --exclude='glb:**/.temp-files/**' or --exclude='(^|/)\\.temp-files(/|$)')"` //nolint:lll
	ListTemplates          bool     `arg:"--list-template"            help:"Lists all default templates in the binary"`
	OutputTemplate         string   `arg:"--out-template"             help:"outputs a template to stdout"`
	PackageDoc             []string `arg:"-d,separate"                help:"set relative package search filepaths for package documentation"                                          placeholder:"FILEPATH"`
	TemplateDir            string   `                                 help:"Loads template files *.gtpl from a directory, use --list to get valid names of templates"`
	TypeLinks              string   `arg:"--type-links"               help:"Controls type reference linking: disabled, internal, or external (default disabled)"`
	Concatenation          string   `arg:"--concatenation"            help:"Controls doc comment concatenation: none or full (default none)"                                                                 default:"none"`
	Highlighter            string   `arg:"--highlighter"              help:"Source code highlighter to use; available: highlightjs, goasciidoc (custom highlightjs)"                                         default:"highlightjs"`
	Render                 []string `arg:"--render,separate"          help:"Controls what examples to render for structs: struct-json, struct-yaml (can specify multiple)"`
	BuildTag               []string `arg:"--build-tag,separate"       help:"Build tags to include when parsing (can specify multiple, e.g., --build-tag=integration --build-tag=dev)" placeholder:"TAG"`
	AllBuildTags           bool     `arg:"--all-build-tags"           help:"Auto-discover and include all build tags found in source files"`
	IgnoreMarkdownHeadings bool     `arg:"--ignore-markdown-headings" help:"Replace markdown headings (#, ##, etc.) in comments with their text content"`
	SubModule              string   `arg:"--sub-module"               help:"Submodule processing mode: none, single, or separate (default none)"                                                             default:"none"`
	PackageMode            string   `arg:"--package-mode"             help:"Package-level rendering mode: none, include, or link (default none)"                                                             default:"none"`
}

func (args) Version() string {
	return "goasciidoc v0.6.0"
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
		Outfile(args.Out)

	if args.Debug {
		p.Debug(true)
	}

	if len(args.Excludes) > 0 {
		p.Excludes(args.Excludes...)
	}

	// Handle workspace and sub-module configuration
	subModuleMode, err := parseSubModuleMode(args.SubModule)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Handle package-level rendering mode
	packageMode, err := parsePackageMode(args.PackageMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	p.PackageMode(packageMode)

	// Determine search path
	searchPath := args.Module
	if searchPath == "" && len(args.Paths) > 0 {
		searchPath = args.Paths[0]
	}
	if searchPath == "" {
		searchPath = "."
	}

	// Discover module/workspace if not explicitly specified
	if args.Module == "" {
		workspace, module, err := goparser.FindModuleOrWorkspace(searchPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find workspace or module: %v\n", err)
			fmt.Fprintf(
				os.Stderr,
				"Make sure you are in a directory with a go.mod or go.work file, or use --module to specify the path.\n",
			)
			os.Exit(1)
		}

		// If sub-module mode is enabled, handle workspace/multiple modules
		if subModuleMode != asciidoc.SubModuleNone {
			if workspace != nil {
				p.Workspace(workspace).SubModule(subModuleMode)
			} else if module != nil {
				// Single module found, but sub-module mode requested
				// Try to discover submodules recursively
				modules, err := goparser.FindAllModules(module.Base)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to find submodules: %v\n", err)
					os.Exit(1)
				}

				if len(modules) > 1 {
					// Found multiple modules - create synthetic workspace
					workspace := &goparser.GoWorkspace{
						Base:      module.Base,
						Modules:   modules,
						ModuleMap: make(map[string]*goparser.GoModule),
					}
					for _, mod := range modules {
						workspace.ModuleMap[mod.Name] = mod
					}
					p.Workspace(workspace).SubModule(subModuleMode)
				} else {
					// Only one module found
					p.Module(module.FilePath)
				}
			}
		} else {
			// Single module mode (default) - use discovered module
			if module != nil {
				p.Module(module.FilePath)
			} else if workspace != nil {
				// Found workspace but single mode - use first module
				if len(workspace.Modules) > 0 {
					p.Module(workspace.Modules[0].FilePath)
				}
			}
		}
	} else {
		// Module path explicitly specified - use it directly
		p.Module(args.Module)
	}

	p.Include(args.Paths...).
		IndexConfig(args.IndexConfig)

	if mode, err := parseTypeLinks(args.TypeLinks); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		p.TypeLinks(mode)
	}

	if mode, err := parseConcatenation(args.Concatenation); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		p.Concatenation(mode)
	}

	if hl, err := parseHighlighter(args.Highlighter); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		switch hl {
		case "highlightjs", "none":
			p.Highlighter("highlightjs").
				SignatureStyle("source")
		case "goasciidoc":
			p.Highlighter("highlightjs").
				SignatureStyle("goasciidoc")
		}
	}

	if renderOpts, err := parseRenderOptions(args.Render); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		p.RenderOptions(renderOpts)
	}

	p.Override(string(asciidoc.ConstDeclarationTemplate), templateConstAssignment)
	p.Override(string(asciidoc.ConstDeclarationsTemplate), templateConstAssignments)
	p.Override(string(asciidoc.FunctionTemplate), templateFunction)
	p.Override(string(asciidoc.FunctionsTemplate), templateFunctions)
	p.Override(string(asciidoc.ImportTemplate), templateImports)
	// Append module template to index template so it can be used with ExecuteTemplate
	p.Override(string(asciidoc.IndexTemplate), templateIndex+"\n"+templateModule)
	p.Override(string(asciidoc.InterfaceTemplate), templateInterface)
	p.Override(string(asciidoc.InterfacesTemplate), templateInterfaces)
	p.Override(string(asciidoc.PackageTemplate), templatePackage)
	p.Override(string(asciidoc.ReceiversTemplate), templateReceivers)
	p.Override(string(asciidoc.StructTemplate), templateStruct)
	p.Override(string(asciidoc.StructsTemplate), templateStructs)
	p.Override(string(asciidoc.CustomFuncTypeDefTemplate), templateCustomFuncDefintion)
	p.Override(string(asciidoc.CustomFuncTypeDefsTemplate), templateCustomFuncDefinitions)
	p.Override(string(asciidoc.CustomVarTypeDefTemplate), templateCustomTypeDefintion)
	p.Override(string(asciidoc.CustomVarTypeDefsTemplate), templateCustomTypeDefinitions)
	p.Override(string(asciidoc.VarDeclarationTemplate), templateVarAssignment)
	p.Override(string(asciidoc.VarDeclarationsTemplate), templateVarAssignments)

	p.EnableMacro()

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

	if len(args.BuildTag) > 0 {
		p.BuildTags(args.BuildTag...)
	}

	if args.AllBuildTags {
		p.AllBuildTags(true)
	}

	if args.IgnoreMarkdownHeadings {
		p.IgnoreMarkdownHeadings(true)
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

func parseSubModuleMode(value string) (asciidoc.SubModuleMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none", "":
		return asciidoc.SubModuleNone, nil
	case "single", "merged":
		return asciidoc.SubModuleSingle, nil
	case "separate", "split":
		return asciidoc.SubModuleSeparate, nil
	default:
		return asciidoc.SubModuleNone, fmt.Errorf(
			"unknown --sub-module mode %q (valid: none, single, separate)",
			value,
		)
	}
}

func parsePackageMode(value string) (asciidoc.PackageMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none", "":
		return asciidoc.PackageModeNone, nil
	case "include":
		return asciidoc.PackageModeInclude, nil
	case "link":
		return asciidoc.PackageModeLink, nil
	default:
		return asciidoc.PackageModeNone, fmt.Errorf(
			"unknown --package-mode %q (valid: none, include, link)",
			value,
		)
	}
}

func parseTypeLinks(value string) (asciidoc.TypeLinkMode, error) {
	if value == "" {
		return asciidoc.TypeLinksDisabled, nil
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "disabled", "off", "none":
		return asciidoc.TypeLinksDisabled, nil
	case "internal", "internal-only":
		return asciidoc.TypeLinksInternal, nil
	case "external", "internal-external", "all":
		return asciidoc.TypeLinksInternalExternal, nil
	default:
		return asciidoc.TypeLinksDisabled, fmt.Errorf(
			"unknown --type-links mode %q (valid: disabled, internal, external)",
			value,
		)
	}
}

func parseConcatenation(value string) (goparser.DocConcatenationMode, error) {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" || v == "none" {
		return goparser.DocConcatenationNone, nil
	}
	if v == "full" {
		return goparser.DocConcatenationFull, nil
	}
	return goparser.DocConcatenationNone, fmt.Errorf(
		"unknown --concatenation mode %q (valid: none, full)",
		value,
	)
}

func parseHighlighter(value string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(value))
	if v == "" {
		return "highlightjs", nil
	}

	switch v {
	case "highlight", "highlightjs", "highlight.js":
		return "highlightjs", nil
	case "goasciidoc":
		return "goasciidoc", nil
	case "none", "off":
		return "none", nil
	default:
		return "", fmt.Errorf(
			"unknown --highlighter %q (available: highlightjs, goasciidoc, none)",
			value,
		)
	}
}

func parseRenderOptions(values []string) (map[string]bool, error) {
	renderOpts := make(map[string]bool)

	for _, v := range values {
		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "struct-json":
			renderOpts["struct-json"] = true
		case "struct-yaml":
			renderOpts["struct-yaml"] = true
		default:
			return nil, fmt.Errorf(
				"unknown --render option %q (valid: struct-json, struct-yaml)",
				v,
			)
		}
	}

	return renderOpts, nil
}
