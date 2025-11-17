package goparser

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentParseSingleFile tests that multiple goroutines can safely
// parse different files concurrently without data races.
func TestConcurrentParseSingleFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Create test files
	writeFile := func(name, contents string) string {
		path := filepath.Join(root, name)
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
		return path
	}

	file1 := writeFile("file1.go", `package test
type Foo struct {
	Name string
}
`)

	file2 := writeFile("file2.go", `package test
type Bar struct {
	Value int
}
`)

	file3 := writeFile("file3.go", `package test
type Baz struct {
	ID uint64
}
`)

	files := []string{file1, file2, file3}

	// Parse files concurrently
	var wg sync.WaitGroup
	results := make([]*GoFile, len(files))
	errors := make([]error, len(files))

	for i, f := range files {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			results[idx], errors[idx] = ParseSingleFile(nil, path)
		}(i, f)
	}

	wg.Wait()

	// Verify all parses succeeded
	for i, err := range errors {
		require.NoError(t, err, "file %d failed to parse", i)
	}

	// Verify results
	assert.Len(t, results[0].Structs, 1)
	assert.Equal(t, "Foo", results[0].Structs[0].Name)

	assert.Len(t, results[1].Structs, 1)
	assert.Equal(t, "Bar", results[1].Structs[0].Name)

	assert.Len(t, results[2].Structs, 1)
	assert.Equal(t, "Baz", results[2].Structs[0].Name)
}

// TestConcurrentParseWithDifferentConfigs tests that multiple goroutines can
// safely parse files with different configurations concurrently.
func TestConcurrentParseWithDifferentConfigs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile := func(name, contents string) string {
		path := filepath.Join(root, name)
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
		return path
	}

	file := writeFile("test.go", `package test
// First comment
// Second comment
type Example struct {}
`)

	// Parse the same file concurrently with different doc concatenation modes
	var wg sync.WaitGroup
	configs := []DocConcatenationMode{
		DocConcatenationNone,
		DocConcatenationFull,
	}

	results := make([]*GoFile, len(configs))
	errors := make([]error, len(configs))

	for i, mode := range configs {
		wg.Add(1)
		go func(idx int, docMode DocConcatenationMode) {
			defer wg.Done()
			cfg := ParseConfig{
				DocConcatenation: docMode,
			}
			results[idx], errors[idx] = parseSingleFileWithConfig(cfg, file)
		}(i, mode)
	}

	wg.Wait()

	// Verify all parses succeeded
	for i, err := range errors {
		require.NoError(t, err, "config %d failed to parse", i)
	}

	// Verify results are different based on doc concatenation mode
	assert.NotNil(t, results[0])
	assert.NotNil(t, results[1])
}

// TestConcurrentParseInlineFile tests concurrent parsing of inline code strings.
func TestConcurrentParseInlineFile(t *testing.T) {
	t.Parallel()

	codes := []string{
		`package alpha
type Alpha struct { A int }`,
		`package beta
type Beta struct { B string }`,
		`package gamma
type Gamma struct { C bool }`,
	}

	var wg sync.WaitGroup
	results := make([]*GoFile, len(codes))
	errors := make([]error, len(codes))

	for i, code := range codes {
		wg.Add(1)
		go func(idx int, src string) {
			defer wg.Done()
			results[idx], errors[idx] = ParseInlineFile(nil, "test.go", src)
		}(i, code)
	}

	wg.Wait()

	// Verify all parses succeeded
	for i, err := range errors {
		require.NoError(t, err, "inline parse %d failed", i)
	}

	// Verify each result
	assert.Equal(t, "alpha", results[0].Package)
	assert.Equal(t, "Alpha", results[0].Structs[0].Name)

	assert.Equal(t, "beta", results[1].Package)
	assert.Equal(t, "Beta", results[1].Structs[0].Name)

	assert.Equal(t, "gamma", results[2].Package)
	assert.Equal(t, "Gamma", results[2].Structs[0].Name)
}

// TestConcurrentResolverCreation tests that multiple resolvers can be
// created concurrently.
func TestConcurrentResolverCreation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile := func(name, contents string) {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	}

	writeFile("go.mod", "module example.com/test\n\ngo 1.24\n")
	writeFile("main.go", `package main
type Main struct{}
`)

	var wg sync.WaitGroup
	numResolvers := 10
	resolvers := make([]Resolver, numResolvers)
	errors := make([]error, numResolvers)

	for i := 0; i < numResolvers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resolvers[idx], errors[idx] = NewResolver(ParseConfig{}, root)
		}(i)
	}

	wg.Wait()

	// Verify all resolvers were created successfully
	for i, err := range errors {
		require.NoError(t, err, "resolver %d failed to create", i)
		assert.NotNil(t, resolvers[i])
	}
}

// TestMassiveConcurrentParsing tests the parser under heavy concurrent load.
func TestMassiveConcurrentParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping massive concurrency test in short mode")
	}

	t.Parallel()

	root := t.TempDir()

	// Create a test file
	testFile := filepath.Join(root, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte(`package test
type Example struct {
	Field string
}

func (e *Example) Method() string {
	return e.Field
}
`), 0o644))

	// Parse the same file 100 times concurrently
	var wg sync.WaitGroup
	numGoroutines := 100
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			goFile, err := ParseSingleFile(nil, testFile)
			if err == nil && goFile != nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// All parses should succeed
	assert.Equal(t, numGoroutines, successCount, "all concurrent parses should succeed")
}
