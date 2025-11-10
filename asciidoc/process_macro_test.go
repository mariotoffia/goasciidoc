package asciidoc

import (
	"strings"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// TestIgnoreMarkdownHeadings tests the markdown heading removal functionality
func TestIgnoreMarkdownHeadings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		enabled  bool
	}{
		{
			name:     "Single hash heading",
			input:    "# Main Title\nSome content",
			expected: "Main Title\nSome content",
			enabled:  true,
		},
		{
			name:     "Double hash heading",
			input:    "## Subsection\nMore content",
			expected: "Subsection\nMore content",
			enabled:  true,
		},
		{
			name:     "Triple hash heading",
			input:    "### Sub-subsection\nEven more content",
			expected: "Sub-subsection\nEven more content",
			enabled:  true,
		},
		{
			name:     "Multiple headings",
			input:    "# Title\n\n## Section\n\n### Subsection\n\nContent here",
			expected: "Title\n\nSection\n\nSubsection\n\nContent here",
			enabled:  true,
		},
		{
			name:     "Heading with special characters",
			input:    "## API Reference: GetUser()\nDescription here",
			expected: "API Reference: GetUser()\nDescription here",
			enabled:  true,
		},
		{
			name:     "Six level heading",
			input:    "###### Deep heading\nContent",
			expected: "Deep heading\nContent",
			enabled:  true,
		},
		{
			name:     "Heading with multiple spaces",
			input:    "##    Multiple Spaces    \nContent",
			expected: "Multiple Spaces\nContent",
			enabled:  true,
		},
		{
			name:     "Empty heading",
			input:    "## \nContent",
			expected: "\nContent",
			enabled:  true,
		},
		{
			name:     "Heading in middle of text",
			input:    "Some text\n## Heading\nMore text",
			expected: "Some text\nHeading\nMore text",
			enabled:  true,
		},
		{
			name:     "Disabled - should not remove headings",
			input:    "# Title\n## Section\nContent",
			expected: "# Title\n## Section\nContent",
			enabled:  false,
		},
		{
			name:     "No headings",
			input:    "Just regular text\nWith multiple lines\nNo headings here",
			expected: "Just regular text\nWith multiple lines\nNo headings here",
			enabled:  true,
		},
		{
			name:     "Hash in middle of line should not be removed",
			input:    "This is # not a heading because it's mid-sentence",
			expected: "This is # not a heading because it's mid-sentence",
			enabled:  true,
		},
		{
			name:     "Heading at start of second line",
			input:    "First line\n## Second line is heading",
			expected: "First line\nSecond line is heading",
			enabled:  true,
		},
		{
			name:     "Heading with trailing content on same line",
			input:    "### Configuration Options\nOption 1: value",
			expected: "Configuration Options\nOption 1: value",
			enabled:  true,
		},
		{
			name:     "Mixed content with macros",
			input:    "# Introduction\n\nSome content\n\n## Details",
			expected: "Introduction\n\nSome content\n\nDetails",
			enabled:  true,
		},
		{
			name:     "Heading with leading spaces",
			input:    "  ## Indented Heading\nContent",
			expected: "Indented Heading\nContent",
			enabled:  true,
		},
		{
			name:     "Heading with leading tab",
			input:    "\t### Tab Heading\nContent",
			expected: "Tab Heading\nContent",
			enabled:  true,
		},
		{
			name:     "Heading without text",
			input:    "##\nContent",
			expected: "\nContent",
			enabled:  true,
		},
		{
			name:     "Heading with only spaces after markers",
			input:    "###   \nContent",
			expected: "\nContent",
			enabled:  true,
		},
		{
			name:     "Mixed leading whitespace",
			input:    " \t # Title\nContent",
			expected: "Title\nContent",
			enabled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a ParseConfig with IgnoreMarkdownHeadings setting
			config := goparser.ParseConfig{
				IgnoreMarkdownHeadings: tt.enabled,
			}

			// Create a mock package with the test input as documentation
			pkg := &goparser.GoPackage{
				GoFile: goparser.GoFile{
					Doc: tt.input,
					Module: &goparser.GoModule{
						Base: "/test/path",
					},
				},
			}

			// Create a no-op next function
			next := func(pkg *goparser.GoPackage) error {
				return nil
			}

			// Get the process function and execute it
			processFunc := getProcessMacroFunc(config, next)
			err := processFunc(pkg)
			if err != nil {
				t.Fatalf("processFunc returned error: %v", err)
			}

			// Verify the result
			if pkg.Doc != tt.expected {
				t.Errorf("Expected:\n%q\n\nGot:\n%q", tt.expected, pkg.Doc)
			}
		})
	}
}

// TestIgnoreMarkdownHeadingsOnStructFields tests markdown heading removal on struct fields
func TestIgnoreMarkdownHeadingsOnStructFields(t *testing.T) {
	config := goparser.ParseConfig{
		IgnoreMarkdownHeadings: true,
	}

	pkg := &goparser.GoPackage{
		GoFile: goparser.GoFile{
			Module: &goparser.GoModule{
				Base: "/test/path",
			},
			Structs: []*goparser.GoStruct{
				{
					Doc: "# Main Struct\nThis is the main struct",
					File: &goparser.GoFile{
						FilePath: "/test/path/file.go",
					},
					Fields: []*goparser.GoField{
						{
							Doc: "## Field One\nFirst field description",
							File: &goparser.GoFile{
								FilePath: "/test/path/file.go",
							},
						},
						{
							Doc: "### Field Two\nSecond field description",
							File: &goparser.GoFile{
								FilePath: "/test/path/file.go",
							},
						},
					},
				},
			},
		},
	}

	next := func(pkg *goparser.GoPackage) error {
		return nil
	}

	processFunc := getProcessMacroFunc(config, next)
	err := processFunc(pkg)
	if err != nil {
		t.Fatalf("processFunc returned error: %v", err)
	}

	// Check struct doc
	expectedStructDoc := "Main Struct\nThis is the main struct"
	if pkg.Structs[0].Doc != expectedStructDoc {
		t.Errorf("Struct doc - Expected:\n%q\n\nGot:\n%q", expectedStructDoc, pkg.Structs[0].Doc)
	}

	// Check field docs
	expectedField1Doc := "Field One\nFirst field description"
	if pkg.Structs[0].Fields[0].Doc != expectedField1Doc {
		t.Errorf("Field 1 doc - Expected:\n%q\n\nGot:\n%q", expectedField1Doc, pkg.Structs[0].Fields[0].Doc)
	}

	expectedField2Doc := "Field Two\nSecond field description"
	if pkg.Structs[0].Fields[1].Doc != expectedField2Doc {
		t.Errorf("Field 2 doc - Expected:\n%q\n\nGot:\n%q", expectedField2Doc, pkg.Structs[0].Fields[1].Doc)
	}
}

// TestIgnoreMarkdownHeadingsOnMethods tests markdown heading removal on interface and struct methods
func TestIgnoreMarkdownHeadingsOnMethods(t *testing.T) {
	config := goparser.ParseConfig{
		IgnoreMarkdownHeadings: true,
	}

	pkg := &goparser.GoPackage{
		GoFile: goparser.GoFile{
			Module: &goparser.GoModule{
				Base: "/test/path",
			},
			Interfaces: []*goparser.GoInterface{
				{
					Doc: "# Main Interface\nInterface description",
					File: &goparser.GoFile{
						FilePath: "/test/path/file.go",
					},
					Methods: []*goparser.GoMethod{
						{
							Doc: "## GetData\nRetrieves data from source",
							File: &goparser.GoFile{
								FilePath: "/test/path/file.go",
							},
						},
					},
				},
			},
			StructMethods: []*goparser.GoStructMethod{
				{
					GoMethod: goparser.GoMethod{
						Doc: "### Process\nProcesses the input",
						File: &goparser.GoFile{
							FilePath: "/test/path/file.go",
						},
					},
				},
			},
		},
	}

	next := func(pkg *goparser.GoPackage) error {
		return nil
	}

	processFunc := getProcessMacroFunc(config, next)
	err := processFunc(pkg)
	if err != nil {
		t.Fatalf("processFunc returned error: %v", err)
	}

	// Check interface doc
	expectedInterfaceDoc := "Main Interface\nInterface description"
	if pkg.Interfaces[0].Doc != expectedInterfaceDoc {
		t.Errorf("Interface doc - Expected:\n%q\n\nGot:\n%q", expectedInterfaceDoc, pkg.Interfaces[0].Doc)
	}

	// Check interface method doc
	expectedMethodDoc := "GetData\nRetrieves data from source"
	if pkg.Interfaces[0].Methods[0].Doc != expectedMethodDoc {
		t.Errorf("Interface method doc - Expected:\n%q\n\nGot:\n%q", expectedMethodDoc, pkg.Interfaces[0].Methods[0].Doc)
	}

	// Check struct method doc
	expectedStructMethodDoc := "Process\nProcesses the input"
	if pkg.StructMethods[0].Doc != expectedStructMethodDoc {
		t.Errorf("Struct method doc - Expected:\n%q\n\nGot:\n%q", expectedStructMethodDoc, pkg.StructMethods[0].Doc)
	}
}

// TestIgnoreMarkdownHeadingsPreservesContent tests that non-heading content is preserved
func TestIgnoreMarkdownHeadingsPreservesContent(t *testing.T) {
	config := goparser.ParseConfig{
		IgnoreMarkdownHeadings: true,
	}

	input := strings.TrimSpace(`
# API Documentation

This package provides utilities for working with APIs.

## Features

- Fast processing
- Easy to use
- Well documented

### Usage

Call the methods as needed. Use # for comments in code.
Note: ## is not removed when it's not at the start of a line.

## Examples

See below for examples.
`)

	expected := strings.TrimSpace(`
API Documentation

This package provides utilities for working with APIs.

Features

- Fast processing
- Easy to use
- Well documented

Usage

Call the methods as needed. Use # for comments in code.
Note: ## is not removed when it's not at the start of a line.

Examples

See below for examples.
`)

	pkg := &goparser.GoPackage{
		GoFile: goparser.GoFile{
			Doc: input,
			Module: &goparser.GoModule{
				Base: "/test/path",
			},
		},
	}

	next := func(pkg *goparser.GoPackage) error {
		return nil
	}

	processFunc := getProcessMacroFunc(config, next)
	err := processFunc(pkg)
	if err != nil {
		t.Fatalf("processFunc returned error: %v", err)
	}

	if pkg.Doc != expected {
		t.Errorf("Expected:\n%q\n\nGot:\n%q", expected, pkg.Doc)
	}
}
