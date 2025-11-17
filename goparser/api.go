package goparser

// This file implements the simplified, modern API for goparser using the
// functional options pattern for better usability and discoverability.

// Option is a functional option for configuring a Parser.
type Option func(*Parser)

// Parser holds parsing configuration and provides methods for parsing Go source code.
// Create a new Parser using NewParser with optional configuration.
//
// Example:
//
//	parser := goparser.NewParser(
//	    goparser.WithModule(mod),
//	    goparser.WithBuildTags("linux"))
//	file, err := parser.ParseFile("main.go")
type Parser struct {
	config ParseConfig
	// virtualPath is used only for ParseCode to specify a virtual file path
	virtualPath string
}

// NewParser creates a new Parser with the specified options.
// If no options are provided, sensible defaults are used.
//
// Example:
//
//	parser := goparser.NewParser(
//	    goparser.WithModule(mod),
//	    goparser.WithTests(true))
func NewParser(opts ...Option) *Parser {
	p := &Parser{
		config: ParseConfig{
			// Sensible defaults
			Test:                   false,
			Internal:               false,
			UnderScore:             false,
			DocConcatenation:       DocConcatenationNone,
			IgnoreMarkdownHeadings: false,
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// ParseFile parses a single Go source file.
//
// Example:
//
//	file, err := parser.ParseFile("main.go")
func (p *Parser) ParseFile(path string) (*GoFile, error) {
	return parseSingleFileWithConfig(p.config, path)
}

// ParseFiles parses multiple Go source files.
// All files are parsed with the same configuration.
//
// Example:
//
//	files, err := parser.ParseFiles("file1.go", "file2.go", "file3.go")
func (p *Parser) ParseFiles(paths ...string) ([]*GoFile, error) {
	return parseFiles(p.config, paths...)
}

// ParseDir recursively parses one or more directories.
// It also accepts individual file paths.
//
// Example:
//
//	files, err := parser.ParseDir("./src", "./cmd")
func (p *Parser) ParseDir(paths ...string) ([]*GoFile, error) {
	return ParseAny(p.config, paths...)
}

// ParseCode parses inline Go source code.
// The virtualPath parameter is optional and can be empty string.
//
// Example:
//
//	code := `package main
//	func main() {}`
//	file, err := parser.ParseCode(code, "main.go")
func (p *Parser) ParseCode(code string, virtualPath string) (*GoFile, error) {
	path := virtualPath
	if path == "" && p.virtualPath != "" {
		path = p.virtualPath
	}
	return ParseInlineFileWithConfig(p.config, path, code)
}

// WalkFiles walks through the specified files one at a time, calling fn for each.
// This is memory efficient for processing large numbers of files.
//
// Example:
//
//	err := parser.WalkFiles(func(file *GoFile) error {
//	    fmt.Println(file.Package)
//	    return nil
//	}, "./src")
func (p *Parser) WalkFiles(fn func(*GoFile) error, paths ...string) error {
	return ParseSingleFileWalker(p.config, fn, paths...)
}

// WalkPackages walks through packages one at a time, calling fn for each.
// Files in the same directory are grouped into a single GoPackage.
// This is memory efficient for processing large codebases.
//
// Example:
//
//	err := parser.WalkPackages(func(pkg *GoPackage) error {
//	    fmt.Printf("Package %s has %d files\n", pkg.Package, len(pkg.Files))
//	    return nil
//	}, "./src")
func (p *Parser) WalkPackages(fn func(*GoPackage) error, paths ...string) error {
	return ParseSinglePackageWalker(p.config, fn, paths...)
}

// Option Functions

// WithModule sets the Go module context for resolving package paths.
//
// Example:
//
//	mod, _ := goparser.NewModule("go.mod")
//	parser := goparser.NewParser(goparser.WithModule(mod))
func WithModule(mod *GoModule) Option {
	return func(p *Parser) {
		p.config.Module = mod
	}
}

// WithBuildTags sets build tags for conditional compilation.
// Each string represents a set of comma-separated tags (e.g., "linux,amd64").
//
// Example:
//
//	parser := goparser.NewParser(
//	    goparser.WithBuildTags("linux,amd64", "integration"))
func WithBuildTags(tags ...string) Option {
	return func(p *Parser) {
		p.config.BuildTags = tags
	}
}

// WithAllBuildTags attempts to discover and load all build tag variants.
//
// Example:
//
//	parser := goparser.NewParser(goparser.WithAllBuildTags())
func WithAllBuildTags() Option {
	return func(p *Parser) {
		p.config.AllBuildTags = true
	}
}

// WithTests includes test files (*_test.go) in parsing.
// By default, test files are excluded.
//
// Example:
//
//	parser := goparser.NewParser(goparser.WithTests(true))
func WithTests(include bool) Option {
	return func(p *Parser) {
		p.config.Test = include
	}
}

// WithInternal includes internal packages in parsing.
// By default, internal packages are excluded.
//
// Example:
//
//	parser := goparser.NewParser(goparser.WithInternal(true))
func WithInternal(include bool) Option {
	return func(p *Parser) {
		p.config.Internal = include
	}
}

// WithUnderScore includes directories starting with underscore.
// By default, underscore directories are excluded.
//
// Example:
//
//	parser := goparser.NewParser(goparser.WithUnderScore(true))
func WithUnderScore(include bool) Option {
	return func(p *Parser) {
		p.config.UnderScore = include
	}
}

// WithDebug sets a debug logging function.
// The function is called with debug messages during parsing.
//
// Example:
//
//	parser := goparser.NewParser(
//	    goparser.WithDebug(func(format string, args ...interface{}) {
//	        log.Printf(format, args...)
//	    }))
func WithDebug(fn DebugFunc) Option {
	return func(p *Parser) {
		p.config.Debug = fn
	}
}

// WithDocConcatenation sets the documentation concatenation mode.
// DocConcatenationFull concatenates doc comments separated by blank lines.
//
// Example:
//
//	parser := goparser.NewParser(
//	    goparser.WithDocConcatenation(goparser.DocConcatenationFull))
func WithDocConcatenation(mode DocConcatenationMode) Option {
	return func(p *Parser) {
		p.config.DocConcatenation = mode
	}
}

// WithIgnoreMarkdownHeadings removes markdown heading markers from documentation.
// When enabled, "# Title" becomes "Title" in doc comments.
//
// Example:
//
//	parser := goparser.NewParser(goparser.WithIgnoreMarkdownHeadings(true))
func WithIgnoreMarkdownHeadings(ignore bool) Option {
	return func(p *Parser) {
		p.config.IgnoreMarkdownHeadings = ignore
	}
}

// WithPath sets a virtual path for inline code parsing.
// This is only used when calling Parser.ParseCode without an explicit path.
//
// Example:
//
//	parser := goparser.NewParser(goparser.WithPath("virtual.go"))
//	file, err := parser.ParseCode(code, "") // Uses "virtual.go"
func WithPath(path string) Option {
	return func(p *Parser) {
		p.virtualPath = path
	}
}

// Convenience Functions (Package Level)

// ParseFile parses a single Go source file with optional configuration.
//
// Example:
//
//	// Simple
//	file, err := goparser.ParseFile("main.go")
//
//	// With module
//	mod, _ := goparser.NewModule("go.mod")
//	file, err := goparser.ParseFile("main.go", goparser.WithModule(mod))
func ParseFile(path string, opts ...Option) (*GoFile, error) {
	p := NewParser(opts...)
	return p.ParseFile(path)
}

// Note: Package-level ParseFiles convenience function is not provided to avoid
// conflict with the legacy ParseFiles(mod *GoModule, paths ...string) function.
// Use Parser.ParseFiles() instead for multiple files:
//
//	parser := goparser.NewParser(goparser.WithModule(mod))
//	files, err := parser.ParseFiles("file1.go", "file2.go")

// ParseDir recursively parses a directory with optional configuration.
//
// Example:
//
//	// Simple
//	files, err := goparser.ParseDir("./src")
//
//	// With options
//	files, err := goparser.ParseDir("./src",
//	    goparser.WithModule(mod),
//	    goparser.WithTests(true),
//	    goparser.WithBuildTags("linux"))
func ParseDir(path string, opts ...Option) ([]*GoFile, error) {
	p := NewParser(opts...)
	return p.ParseDir(path)
}

// ParseCode parses inline Go source code with optional configuration.
//
// Example:
//
//	code := `package main
//	func main() {}`
//	file, err := goparser.ParseCode(code)
//
//	// With virtual path
//	file, err := goparser.ParseCode(code, goparser.WithPath("main.go"))
func ParseCode(code string, opts ...Option) (*GoFile, error) {
	p := NewParser(opts...)
	return p.ParseCode(code, "")
}
