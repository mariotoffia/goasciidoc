package goparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestFile creates a temp directory with a simple Go test file
func setupTestFile(t *testing.T) (string, string) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.go")
	code := `package testpkg

// Example is a test struct
type Example struct {
	Name string
}

// NewExample creates a new Example
func NewExample(name string) *Example {
	return &Example{Name: name}
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(code), 0644))
	return tmpDir, testFile
}

// setupTestFileWithModule creates a temp directory with a Go file and go.mod
func setupTestFileWithModule(t *testing.T) (string, string, string) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.go")
	modFile := filepath.Join(tmpDir, "go.mod")

	code := `package testpkg

// Example is a test struct
type Example struct {
	Name string
}

// NewExample creates a new Example
func NewExample(name string) *Example {
	return &Example{Name: name}
}
`
	modContent := `module example.com/test

go 1.21
`
	require.NoError(t, os.WriteFile(testFile, []byte(code), 0644))
	require.NoError(t, os.WriteFile(modFile, []byte(modContent), 0644))
	return tmpDir, testFile, modFile
}

// Test simple file parsing without options
func TestParseFile_Simple(t *testing.T) {
	_, testFile := setupTestFile(t)

	file, err := ParseFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, "testpkg", file.Package)
}

// Test file parsing with module option
func TestParseFile_WithModule(t *testing.T) {
	_, testFile, modFile := setupTestFileWithModule(t)

	mod, err := NewModule(modFile)
	require.NoError(t, err)

	file, err := ParseFile(testFile, WithModule(mod))
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, mod, file.Module)
	assert.Equal(t, "testpkg", file.Package)
}

// Test file parsing with multiple options
func TestParseFile_WithMultipleOptions(t *testing.T) {
	_, testFile, modFile := setupTestFileWithModule(t)

	mod, err := NewModule(modFile)
	require.NoError(t, err)

	file, err := ParseFile(testFile,
		WithModule(mod),
		WithBuildTags("linux"),
		WithDebug(func(format string, args ...interface{}) {
			// Debug function registered but may not be called for simple ParseFile
		}))

	require.NoError(t, err)
	require.NotNil(t, file)
	// Note: Debug function is not currently called by ParseFile's internal implementation
	// This test verifies the option is accepted and parsing succeeds
}

// Test Parser.ParseFile method
func TestParser_ParseFile(t *testing.T) {
	_, testFile := setupTestFile(t)

	parser := NewParser()
	file, err := parser.ParseFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, "testpkg", file.Package)
}

// Test Parser can be reused
func TestParser_Reusable(t *testing.T) {
	_, testFile, modFile := setupTestFileWithModule(t)

	mod, err := NewModule(modFile)
	require.NoError(t, err)

	parser := NewParser(WithModule(mod))

	// Parse first file
	file1, err := parser.ParseFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, file1)

	// Parse same file again with same parser
	file2, err := parser.ParseFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, file2)

	// Both should have same module
	assert.Equal(t, file1.Module, file2.Module)
}

// Test Parser.ParseFiles method with multiple files
func TestParser_ParseFiles(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	code1 := `package testpkg
type Foo struct {
	Name string
}`
	code2 := `package testpkg
type Bar struct {
	Value int
}`

	require.NoError(t, os.WriteFile(file1, []byte(code1), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(code2), 0644))

	parser := NewParser()
	files, err := parser.ParseFiles(file1, file2)
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "testpkg", files[0].Package)
	assert.Equal(t, "testpkg", files[1].Package)
}

// Test ParseDir convenience function
func TestParseDir_Simple(t *testing.T) {
	tmpDir, _ := setupTestFile(t)

	files, err := ParseDir(tmpDir)
	require.NoError(t, err)
	require.NotEmpty(t, files)
}

// Test ParseDir with options
func TestParseDir_WithOptions(t *testing.T) {
	tmpDir, _, modFile := setupTestFileWithModule(t)

	mod, err := NewModule(modFile)
	require.NoError(t, err)

	files, err := ParseDir(tmpDir,
		WithModule(mod),
		WithInternal(false))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// Verify all files have the module
	for _, f := range files {
		assert.Equal(t, mod, f.Module)
	}
}

// Test Parser.ParseDir method
func TestParser_ParseDir(t *testing.T) {
	tmpDir, _ := setupTestFile(t)

	parser := NewParser()
	files, err := parser.ParseDir(tmpDir)
	require.NoError(t, err)
	require.NotEmpty(t, files)
}

// Test ParseCode convenience function
func TestParseCode_Simple(t *testing.T) {
	code := `package main
// Hello returns a greeting
func Hello() string {
	return "world"
}`

	file, err := ParseCode(code)
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, "main", file.Package)
}

// Test ParseCode with virtual path
func TestParseCode_WithPath(t *testing.T) {
	code := `package testpkg
type MyStruct struct {}`

	file, err := ParseCode(code, WithPath("virtual.go"))
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, "testpkg", file.Package)
}

// Test Parser.ParseCode method
func TestParser_ParseCode(t *testing.T) {
	code := `package example
const Version = "1.0.0"`

	parser := NewParser()
	file, err := parser.ParseCode(code, "version.go")
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, "example", file.Package)
}

// Test Parser.ParseCode uses virtualPath from WithPath option
func TestParser_ParseCode_UsesVirtualPath(t *testing.T) {
	code := `package example
type Data struct {}`

	parser := NewParser(WithPath("data.go"))
	file, err := parser.ParseCode(code, "")
	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, "example", file.Package)
}

// Test WalkFiles method
func TestParser_WalkFiles(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	code := `package test
type Example struct {}`
	require.NoError(t, os.WriteFile(testFile, []byte(code), 0644))

	parser := NewParser()
	var count int
	var seenPackages []string

	err := parser.WalkFiles(func(file *GoFile) error {
		count++
		seenPackages = append(seenPackages, file.Package)
		return nil
	}, testFile)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, []string{"test"}, seenPackages)
}

// Test WalkPackages method
func TestParser_WalkPackages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files in same package
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")

	require.NoError(t, os.WriteFile(file1, []byte("package testpkg\ntype Type1 struct {}"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("package testpkg\ntype Type2 struct {}"), 0644))

	parser := NewParser()
	var count int
	var packageNames []string

	err := parser.WalkPackages(func(pkg *GoPackage) error {
		count++
		packageNames = append(packageNames, pkg.Package)
		return nil
	}, tmpDir)

	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should have one package")
	assert.Equal(t, []string{"testpkg"}, packageNames)
}

// Test all option functions
func TestOptions_AllApplied(t *testing.T) {
	_, testFile, modFile := setupTestFileWithModule(t)

	mod, err := NewModule(modFile)
	require.NoError(t, err)

	parser := NewParser(
		WithModule(mod),
		WithBuildTags("linux", "amd64"),
		WithAllBuildTags(),
		WithTests(true),
		WithInternal(true),
		WithUnderScore(true),
		WithDebug(func(format string, args ...interface{}) {
			// Debug function registered but may not be called for simple ParseFile
		}),
		WithDocConcatenation(DocConcatenationFull),
		WithIgnoreMarkdownHeadings(true),
		WithPath("virtual.go"),
	)

	assert.Equal(t, mod, parser.config.Module)
	assert.Equal(t, []string{"linux", "amd64"}, parser.config.BuildTags)
	assert.True(t, parser.config.AllBuildTags)
	assert.True(t, parser.config.Test)
	assert.True(t, parser.config.Internal)
	assert.True(t, parser.config.UnderScore)
	assert.NotNil(t, parser.config.Debug)
	assert.Equal(t, DocConcatenationFull, parser.config.DocConcatenation)
	assert.True(t, parser.config.IgnoreMarkdownHeadings)
	assert.Equal(t, "virtual.go", parser.virtualPath)

	// Verify parser can parse successfully with all options
	file, err := parser.ParseFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, file)
}

// Test that new API produces same results as old API
func TestNewAPI_EquivalentToOldAPI(t *testing.T) {
	_, testFile, modFile := setupTestFileWithModule(t)

	mod, err := NewModule(modFile)
	require.NoError(t, err)

	// Old API
	oldFile, err1 := ParseSingleFile(mod, testFile)
	require.NoError(t, err1)

	// New API
	newFile, err2 := ParseFile(testFile, WithModule(mod))
	require.NoError(t, err2)

	// Should produce identical key results
	assert.Equal(t, oldFile.Package, newFile.Package)
	assert.Equal(t, oldFile.Module.Name, newFile.Module.Name)
	assert.Equal(t, len(oldFile.Structs), len(newFile.Structs))
	assert.Equal(t, len(oldFile.Interfaces), len(newFile.Interfaces))
	assert.Equal(t, len(oldFile.CustomFuncs), len(newFile.CustomFuncs))
}

// Test NewParser with no options uses sensible defaults
func TestNewParser_Defaults(t *testing.T) {
	parser := NewParser()

	assert.Nil(t, parser.config.Module)
	assert.False(t, parser.config.Test)
	assert.False(t, parser.config.Internal)
	assert.False(t, parser.config.UnderScore)
	assert.Nil(t, parser.config.Debug)
	assert.Equal(t, DocConcatenationNone, parser.config.DocConcatenation)
	assert.False(t, parser.config.IgnoreMarkdownHeadings)
	assert.Empty(t, parser.config.BuildTags)
	assert.False(t, parser.config.AllBuildTags)
}

// Test error handling
func TestParseFile_NonExistentFile(t *testing.T) {
	file, err := ParseFile("nonexistent.go")
	assert.Error(t, err)
	assert.Nil(t, file)
}

// Test ParseCode with invalid syntax
func TestParseCode_InvalidSyntax(t *testing.T) {
	code := `package main
func broken( {{{ syntax error`

	file, err := ParseCode(code)
	assert.Error(t, err)
	assert.Nil(t, file)
}

// Benchmark new API vs old API
func BenchmarkNewAPI_ParseFile(b *testing.B) {
	// Setup temp file once for all iterations
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "sample.go")
	code := `package testpkg

// Example is a test struct
type Example struct {
	Name string
}

// NewExample creates a new Example
func NewExample(name string) *Example {
	return &Example{Name: name}
}
`
	if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseFile(testFile)
	}
}

func BenchmarkOldAPI_ParseSingleFile(b *testing.B) {
	// Setup temp file once for all iterations
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "sample.go")
	code := `package testpkg

// Example is a test struct
type Example struct {
	Name string
}

// NewExample creates a new Example
func NewExample(name string) *Example {
	return &Example{Name: name}
}
`
	if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseSingleFile(nil, testFile)
	}
}

func BenchmarkParser_Reusable(b *testing.B) {
	// Setup temp file once for all iterations
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "sample.go")
	code := `package testpkg

// Example is a test struct
type Example struct {
	Name string
}

// NewExample creates a new Example
func NewExample(name string) *Example {
	return &Example{Name: name}
}
`
	if err := os.WriteFile(testFile, []byte(code), 0644); err != nil {
		b.Fatal(err)
	}

	parser := NewParser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseFile(testFile)
	}
}
