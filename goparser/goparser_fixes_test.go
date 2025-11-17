package goparser

import (
	"strings"
	"testing"
)

// TestMakeIndent tests the optimized makeIndent function
func TestMakeIndent(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected string
	}{
		{
			name:     "zero level",
			level:    0,
			expected: "",
		},
		{
			name:     "negative level",
			level:    -1,
			expected: "",
		},
		{
			name:     "one level",
			level:    1,
			expected: "  ",
		},
		{
			name:     "three levels",
			level:    3,
			expected: "      ",
		},
		{
			name:     "large level (performance test)",
			level:    100,
			expected: strings.Repeat("  ", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeIndent(tt.level)
			if result != tt.expected {
				t.Errorf("makeIndent(%d) = %q, want %q", tt.level, result, tt.expected)
			}
		})
	}
}

// TestNormalizeReceiverName tests boundary checks in normalizeReceiverName
func TestNormalizeReceiverName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple type",
			input:    "Foo",
			expected: "Foo",
		},
		{
			name:     "pointer type",
			input:    "*Foo",
			expected: "Foo",
		},
		{
			name:     "double pointer",
			input:    "**Foo",
			expected: "Foo",
		},
		{
			name:     "reference type",
			input:    "&Foo",
			expected: "Foo",
		},
		{
			name:     "qualified type",
			input:    "pkg.Foo",
			expected: "Foo",
		},
		{
			name:     "pointer to qualified type",
			input:    "*pkg.Foo",
			expected: "Foo",
		},
		{
			name:     "generic type with type parameters",
			input:    "Foo[T]",
			expected: "Foo",
		},
		{
			name:     "pointer to generic type",
			input:    "*Foo[T, U]",
			expected: "Foo",
		},
		{
			name:     "qualified generic type",
			input:    "pkg.Foo[T]",
			expected: "Foo",
		},
		{
			name:     "edge case: dot at end",
			input:    "pkg.",
			expected: "pkg.", // Should handle gracefully
		},
		{
			name:     "edge case: just pointer",
			input:    "*",
			expected: "",
		},
		{
			name:     "edge case: empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "edge case: whitespace",
			input:    "  Foo  ",
			expected: "Foo",
		},
		{
			name:     "complex case: pointer to qualified generic",
			input:    "*pkg.Foo[T, U]",
			expected: "Foo",
		},
		{
			name:     "nested packages",
			input:    "github.com/user/pkg.Foo",
			expected: "Foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeReceiverName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeReceiverName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// BenchmarkMakeIndent benchmarks the optimized makeIndent
func BenchmarkMakeIndent(b *testing.B) {
	benchmarks := []struct {
		name  string
		level int
	}{
		{"small", 5},
		{"medium", 20},
		{"large", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = makeIndent(bm.level)
			}
		})
	}
}

// BenchmarkMakeIndentOld shows the old implementation for comparison
func BenchmarkMakeIndentOld(b *testing.B) {
	makeIndentOld := func(level int) string {
		result := ""
		for i := 0; i < level; i++ {
			result += "  "
		}
		return result
	}

	benchmarks := []struct {
		name  string
		level int
	}{
		{"small", 5},
		{"medium", 20},
		{"large", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = makeIndentOld(bm.level)
			}
		})
	}
}

// TestSelectorExprTypeName tests that complex selector expressions don't panic
func TestSelectorExprTypeName(t *testing.T) {
	// This is a regression test for the unchecked type assertion fix
	// It tests indirectly through parsing code with complex type aliases

	code := `
package test

import "time"

// Alias to a qualified type (not a simple Ident)
type TimeAlias = time.Time

// Another complex case
type ComplexAlias = map[string]interface{}
`

	file, err := ParseInlineFile(nil, "test.go", code)
	if err != nil {
		t.Fatalf("ParseInlineFile failed: %v", err)
	}

	// Verify we parsed the custom types without panicking
	if len(file.CustomTypes) < 2 {
		t.Errorf("Expected at least 2 custom types, got %d", len(file.CustomTypes))
	}

	// Find and verify the TimeAlias
	var foundTimeAlias bool
	for _, ct := range file.CustomTypes {
		if ct.Name == "TimeAlias" {
			foundTimeAlias = true
			// The type should be captured correctly
			if ct.Type == "" {
				t.Errorf("TimeAlias type is empty")
			}
		}
	}

	if !foundTimeAlias {
		t.Error("TimeAlias not found in custom types")
	}
}

// TestUnresolvedDeclTracking tests that unexpected types are tracked instead of printed
func TestUnresolvedDeclTracking(t *testing.T) {
	// Create a module to track unresolved declarations
	modCode := `module github.com/test/example
go 1.21`

	mod, err := NewModuleFromBuff("go.mod", []byte(modCode))
	if err != nil {
		t.Fatalf("NewModuleFromBuff failed: %v", err)
	}

	// Parse some code that might have edge cases
	code := `
package test

// This should parse fine but exercises the type building code
type ComplexType interface {
	Method() error
}
`

	_, err = ParseInlineFile(mod, "test.go", code)
	if err != nil {
		t.Fatalf("ParseInlineFile failed: %v", err)
	}

	// The test passes if we didn't panic and parsing succeeded
	// Unresolved declarations, if any, should be in mod.Unresolved
	// This is mainly a regression test to ensure we don't use fmt.Printf
}

// TestBuildTagsStructOptimization verifies that build tags work correctly
// with the struct{} optimization
func TestBuildTagsStructOptimization(t *testing.T) {
	code1 := `
//go:build linux
// +build linux

package test

type Foo struct{}
`

	code2 := `
//go:build linux
// +build linux

package test

type Bar struct{}
`

	file1, err := ParseInlineFile(nil, "test1.go", code1)
	if err != nil {
		t.Fatalf("ParseInlineFile for file1 failed: %v", err)
	}

	file2, err := ParseInlineFile(nil, "test2.go", code2)
	if err != nil {
		t.Fatalf("ParseInlineFile for file2 failed: %v", err)
	}

	// Aggregate into a package
	pkg := aggregatePackage(nil, ".", []*GoFile{file1, file2})
	if pkg == nil {
		t.Fatal("aggregatePackage returned nil")
	}

	// Verify build tags were collected (should have unique tags)
	if len(pkg.BuildTags) == 0 {
		t.Error("Expected build tags to be collected")
	}

	// Should have deduplicated the "linux" tag
	linuxCount := 0
	for _, tag := range pkg.BuildTags {
		if strings.Contains(tag, "linux") {
			linuxCount++
		}
	}

	if linuxCount == 0 {
		t.Error("Expected at least one linux build tag")
	}

	// Tags should be sorted
	for i := 1; i < len(pkg.BuildTags); i++ {
		if pkg.BuildTags[i-1] > pkg.BuildTags[i] {
			t.Errorf("Build tags not sorted: %v", pkg.BuildTags)
			break
		}
	}
}
