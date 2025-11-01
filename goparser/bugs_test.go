package goparser

import (
	"go/parser"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSingleFilePreservesDocComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.go")
	source := "// Package sample doc\npackage sample\n"
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	goFile, err := ParseSingleFile(nil, path)
	require.NoError(t, err)
	assert.Equal(t, "Package sample doc", goFile.Doc)
}

func TestParseHandlesArrayLengthIdentifier(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "array.go")
	source := `package sample

const size = 4

type buffer [size]int
`
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	var goFile *GoFile
	assert.NotPanics(t, func() {
		var err error
		goFile, err = ParseSingleFile(nil, path)
		require.NoError(t, err)
	})

	require.NotNil(t, goFile)
	require.Len(t, goFile.CustomTypes, 1)
	assert.Equal(t, "[size]int", goFile.CustomTypes[0].Type)
}

func TestRenderTypeNameSupportsCompositeTypes(t *testing.T) {
	testCases := []string{
		"[]int",
		"[size][]string",
		"[]map[string]int",
		"chan<- int",
		"struct{A []int}",
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			expr, err := parser.ParseExpr(tc)
			require.NoError(t, err)
			assert.Equal(t, tc, renderTypeName(expr))
		})
	}
}

func TestParseSingleFileDoesNotPanicWhenModuleNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "channel.go")
	source := `package sample

type stream chan<- int
`
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	var goFile *GoFile
	assert.NotPanics(t, func() {
		var err error
		goFile, err = ParseSingleFile(nil, path)
		require.NoError(t, err)
	})

	require.NotNil(t, goFile)
	require.Len(t, goFile.CustomTypes, 1)
	assert.Equal(t, "chan<- int", goFile.CustomTypes[0].Type)
}

func TestGetFilePathsSkipsInternalByDefault(t *testing.T) {
	dir := t.TempDir()

	mainFile := filepath.Join(dir, "main.go")
	require.NoError(t, os.WriteFile(mainFile, []byte("package sample\n"), 0o644))

	internalDir := filepath.Join(dir, "internal")
	require.NoError(t, os.MkdirAll(internalDir, 0o755))
	internalFile := filepath.Join(internalDir, "hidden.go")
	require.NoError(t, os.WriteFile(internalFile, []byte("package internal\n"), 0o644))

	files, err := GetFilePaths(ParseConfig{}, dir)
	require.NoError(t, err)
	assert.Equal(t, []string{mainFile}, files)

	files, err = GetFilePaths(ParseConfig{Internal: true}, dir)
	require.NoError(t, err)
	assert.Equal(t, []string{internalFile, mainFile}, files)
}

func TestFindMethodsByReceiverHandlesTypeParameters(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "generic.go")
	source := `package sample

type Set[T any] struct{}

func (s *Set[T]) Add() {}
`
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	goFile, err := ParseSingleFile(nil, path)
	require.NoError(t, err)

	methods := goFile.FindMethodsByReceiver("Set")
	require.Len(t, methods, 1)
	assert.Equal(t, "Add", methods[0].Name)
}

func TestImportPathUsesModuleInformation(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg"), 0o755))
	path := filepath.Join(dir, "pkg", "file.go")
	source := "package pkg\n"
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	mod := &GoModule{
		Name: "example.com/app",
		Base: dir,
	}

	goFile, err := ParseSingleFile(mod, path)
	require.NoError(t, err)

	t.Setenv("GOPATH", "")
	importPath, err := goFile.ImportPath()
	require.NoError(t, err)
	assert.Equal(t, "example.com/app/pkg", importPath)
}

func TestImportPathAutoDetectsModule(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg"), 0o755))

	goMod := filepath.Join(dir, "go.mod")
	require.NoError(t, os.WriteFile(goMod, []byte("module example.com/app\n\ngo 1.24\n"), 0o644))

	path := filepath.Join(dir, "pkg", "file.go")
	require.NoError(t, os.WriteFile(path, []byte("package pkg\n"), 0o644))

	goFile, err := ParseSingleFile(nil, path)
	require.NoError(t, err)

	importPath, err := goFile.ImportPath()
	require.NoError(t, err)
	assert.Equal(t, "example.com/app/pkg", importPath)
	assert.NotNil(t, goFile.Module)
	assert.Equal(t, "example.com/app", goFile.Module.Name)
}

func TestImportPathReturnsHelpfulErrorWithoutModule(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "pkg"), 0o755))
	path := filepath.Join(dir, "pkg", "file.go")
	require.NoError(t, os.WriteFile(path, []byte("package pkg\n"), 0o644))

	goFile, err := ParseSingleFile(nil, path)
	require.NoError(t, err)

	_, err = goFile.ImportPath()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "go.mod")
	assert.Contains(t, err.Error(), "ParseConfig.Module")
}

func TestGenericTypeMetadataCaptured(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "generic.go")
	source := `package sample

import "io"

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type List[T any] []T

type Constraint[T any] interface {
	~[]T | *List[T]
	io.Reader
	Do(T) error
}

type Set[T any] struct{}

func Transform[T any](in T) T {
	return in
}

type Mapper[K comparable, V any] func(K) V
`
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	goFile, err := ParseSingleFile(nil, path)
	require.NoError(t, err)

	findStruct := func(name string) *GoStruct {
		for _, s := range goFile.Structs {
			if s.Name == name {
				return s
			}
		}
		return nil
	}

	findInterface := func(name string) *GoInterface {
		for _, i := range goFile.Interfaces {
			if i.Name == name {
				return i
			}
		}
		return nil
	}

	findCustomType := func(name string) *GoCustomType {
		for _, ct := range goFile.CustomTypes {
			if ct.Name == name {
				return ct
			}
		}
		return nil
	}

	findCustomFunc := func(name string) *GoMethod {
		for _, fn := range goFile.CustomFuncs {
			if fn.Name == name {
				return fn
			}
		}
		return nil
	}

	findStructMethod := func(name string) *GoStructMethod {
		for _, m := range goFile.StructMethods {
			if m.Name == name {
				return m
			}
		}
		return nil
	}

	pair := findStruct("Pair")
	require.NotNil(t, pair)
	require.Len(t, pair.TypeParams, 2)
	assert.Equal(t, "K", pair.TypeParams[0].Name)
	assert.Equal(t, "comparable", pair.TypeParams[0].Type)
	assert.Equal(t, "V", pair.TypeParams[1].Name)
	assert.Equal(t, "any", pair.TypeParams[1].Type)

	set := findStruct("Set")
	require.NotNil(t, set)
	require.Len(t, set.TypeParams, 1)
	assert.Equal(t, "T", set.TypeParams[0].Name)

	list := findCustomType("List")
	require.NotNil(t, list)
	require.Len(t, list.TypeParams, 1)
	assert.Equal(t, "T", list.TypeParams[0].Name)
	assert.Equal(t, "any", list.TypeParams[0].Type)
	assert.Equal(t, "[]T", list.Type)

	constraint := findInterface("Constraint")
	require.NotNil(t, constraint)
	require.Len(t, constraint.TypeParams, 1)
	assert.Equal(t, "T", constraint.TypeParams[0].Name)

	typeSet := []string{}
	for _, ts := range constraint.TypeSet {
		typeSet = append(typeSet, ts.Type)
	}
	assert.ElementsMatch(t, []string{"~[]T", "*List[T]", "io.Reader"}, typeSet)
	assert.ElementsMatch(t, []string{"~[]T | *List[T]", "io.Reader"}, constraint.TypeSetDecl)
	require.Len(t, constraint.Methods, 1)
	assert.Equal(t, "Do", constraint.Methods[0].Name)

	mapper := findCustomFunc("Mapper")
	require.NotNil(t, mapper)
	require.Len(t, mapper.TypeParams, 2)
	assert.Equal(t, []string{"K", "V"}, []string{mapper.TypeParams[0].Name, mapper.TypeParams[1].Name})

	transform := findStructMethod("Transform")
	require.NotNil(t, transform)
	require.Len(t, transform.TypeParams, 1)
	assert.Equal(t, "T", transform.TypeParams[0].Name)
	assert.Equal(t, "any", transform.TypeParams[0].Type)
}

func TestInterfaceNestedTypeSet(t *testing.T) {
	source := `package sample

type Reader interface{ Read() }
type Writer interface{ Write() }
type Closer interface{ Close() }
type List[T any] []T

type Combo[T any] interface {
	Reader | (*List[T]) | Closer
	(Writer)
}
`

	goFile, err := ParseInlineFile(nil, "", source)
	require.NoError(t, err)

	var combo *GoInterface
	for _, iface := range goFile.Interfaces {
		if iface.Name == "Combo" {
			combo = iface
			break
		}
	}
	require.NotNil(t, combo, "Combo interface not parsed")

	assert.ElementsMatch(t, []string{"Reader", "*List[T]", "Closer", "Writer"}, collectTypeStrings(combo.TypeSet))
	assert.ElementsMatch(t, []string{"Reader | (*List[T]) | Closer", "(Writer)"}, combo.TypeSetDecl)
}

func collectTypeStrings(types []*GoType) []string {
	result := make([]string, 0, len(types))
	for _, tp := range types {
		if tp != nil {
			result = append(result, tp.Type)
		}
	}
	return result
}
