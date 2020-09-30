// Package goparser was taken from an open source project (https://github.com/zpatrick/go-parser) by zpatrick. Since it seemed
// that he had abandon it, I've integrated it into this project (and extended it).
package goparser

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
)

// ParseSingleFile parses a single file at the same time
func ParseSingleFile(path string) (*GoFile, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	return parseFile(path, nil, file, fset, []*ast.File{file})
}

// ParseFiles parses one or more files
func ParseFiles(paths ...string) ([]*GoFile, error) {

	if len(paths) == 0 {
		return nil, fmt.Errorf("Must specify atleast one path to file to parse")
	}

	files := make([]*ast.File, len(paths))
	fsets := make([]*token.FileSet, len(paths))
	for i, p := range paths {
		// File: A File node represents a Go source file: https://golang.org/pkg/go/ast/#File
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files[i] = file
		fsets[i] = fset
	}

	goFiles := make([]*GoFile, len(paths))
	for i, p := range paths {
		goFile, err := parseFile(p, nil, files[i], fsets[i], files)
		if err != nil {
			return nil, err
		}
		goFiles[i] = goFile
	}
	return goFiles, nil
}

// ParseInlineFile will parse the code provided and have a path of ""
func ParseInlineFile(code string) (*GoFile, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return parseFile("", []byte(code), file, fset, []*ast.File{file})
}

func parseFile(path string, source []byte, file *ast.File, fset *token.FileSet, files []*ast.File) (*GoFile, error) {

	var err error
	if len(source) == 0 {
		source, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	// To import sources from vendor, we use "source" compile
	// https://github.com/golang/go/issues/11415#issuecomment-283445198
	conf := types.Config{Importer: importer.For("source", nil)}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	if _, err = conf.Check(file.Name.Name, fset, files, info); err != nil {
		return nil, err
	}

	goFile := &GoFile{
		Path:    path,
		Doc:     ExtractDocs(file.Doc),
		Package: file.Name.Name,
		Structs: []*GoStruct{},
	}

	// File.Decls: A list of the declarations in the file: https://golang.org/pkg/go/ast/#Decl
	for _, decl := range file.Decls {
		switch declType := decl.(type) {

		// GenDecl: represents an import, constant, type or variable declaration: https://golang.org/pkg/go/ast/#GenDecl
		case *ast.GenDecl:
			genDecl := declType

			// Specs: the Spec type stands for any of *ImportSpec, *ValueSpec, and *TypeSpec: https://golang.org/pkg/go/ast/#Spec
			for _, genSpec := range genDecl.Specs {
				switch genSpecType := genSpec.(type) {

				// TypeSpec: A TypeSpec node represents a type declaration: https://golang.org/pkg/go/ast/#TypeSpec
				case *ast.TypeSpec:
					typeSpec := genSpecType

					// typeSpec.Type: an Expr (expression) node: https://golang.org/pkg/go/ast/#Expr
					switch typeSpecType := typeSpec.Type.(type) {

					// StructType: A StructType node represents a struct type: https://golang.org/pkg/go/ast/#StructType
					case (*ast.StructType):
						structType := typeSpecType
						goStruct := buildGoStruct(source, goFile, info, typeSpec, structType)
						goStruct.Doc = ExtractDocs(declType.Doc)
						goFile.Structs = append(goFile.Structs, goStruct)
					// InterfaceType: An InterfaceType node represents an interface type. https://golang.org/pkg/go/ast/#InterfaceType
					case (*ast.InterfaceType):
						interfaceType := typeSpecType
						goInterface := buildGoInterface(source, goFile, info, typeSpec, interfaceType)
						goInterface.Doc = ExtractDocs(declType.Doc)
						goFile.Interfaces = append(goFile.Interfaces, goInterface)
					// Custom Type declaration
					case (*ast.Ident):
						goCustomType := &GoCustomType{
							Name: genSpecType.Name.Name,
							Type: typeSpecType.Name,
							Doc:  ExtractDocs(declType.Doc),
						}

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					default:
						// a not-implemented typeSpec.Type.(type), ignore
					}
					// ImportSpec: An ImportSpec node represents a single package import. https://golang.org/pkg/go/ast/#ImportSpec
				case *ast.ImportSpec:
					importSpec := genSpec.(*ast.ImportSpec)
					goImport := buildGoImport(importSpec, goFile)
					goFile.Imports = append(goFile.Imports, goImport)
				case *ast.ValueSpec:
					valueSpec := genSpecType

					switch genDecl.Tok {
					case token.VAR:
						goFile.VarAssigments = append(goFile.VarAssigments, buildVarAssignment(genDecl, valueSpec)...)
					case token.CONST:
						goFile.ConstAssignments = append(goFile.ConstAssignments, buildVarAssignment(genDecl, valueSpec)...)
					}
				default:
					// a not-implemented genSpec.(type), ignore
				}
			}
		case *ast.FuncDecl:
			funcDecl := declType
			goStructMethod := buildStructMethod(info, funcDecl, source)
			goFile.StructMethods = append(goFile.StructMethods, goStructMethod)

		default:
			// a not-implemented decl.(type), ignore
		}
	}

	return goFile, nil
}

func buildVarAssignment(genDecl *ast.GenDecl, valueSpec *ast.ValueSpec) []*GoAssignment {

	list := []*GoAssignment{}
	for i := range valueSpec.Names {

		goVarAssignment := &GoAssignment{
			Name: valueSpec.Names[i].Name,
		}

		if genDecl.Doc != nil {
			goVarAssignment.Doc = ExtractDocs(genDecl.Doc)
		}

		if valueSpec.Doc != nil {
			goVarAssignment.Doc = ExtractDocs(valueSpec.Doc)
		}

		list = append(list, goVarAssignment)
	}

	return list
}

// ExtractDocs will extract documentation (if any) from a comment group.
func ExtractDocs(doc *ast.CommentGroup) string {
	return doc.Text()
}

func buildGoImport(spec *ast.ImportSpec, file *GoFile) *GoImport {
	name := ""
	if spec.Name != nil {
		name = spec.Name.Name
	}

	path := ""
	if spec.Path != nil {
		path = spec.Path.Value[1 : len(spec.Path.Value)-1]
	}

	return &GoImport{
		Name: name,
		Path: path,
		File: file,
		Doc:  ExtractDocs(spec.Doc),
	}
}

func buildGoInterface(source []byte, file *GoFile, info *types.Info, typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType) *GoInterface {

	return &GoInterface{
		File:    file,
		Name:    typeSpec.Name.Name,
		Methods: buildMethodList(info, interfaceType.Methods.List, source),
	}

}

func buildMethodList(info *types.Info, fieldList []*ast.Field, source []byte) []*GoMethod {
	methods := []*GoMethod{}

	for _, field := range fieldList {
		name := getNames(field)[0]

		fType, ok := field.Type.(*ast.FuncType)
		if !ok {
			// method was not a function
			continue
		}

		goMethod := &GoMethod{
			Name:    name,
			Params:  buildTypeList(info, fType.Params, source),
			Results: buildTypeList(info, fType.Results, source),
			Doc:     ExtractDocs(field.Doc),
		}

		methods = append(methods, goMethod)
	}

	return methods
}

func buildStructMethod(info *types.Info, funcDecl *ast.FuncDecl, source []byte) *GoStructMethod {

	return &GoStructMethod{
		Receivers: buildReceiverList(info, funcDecl.Recv, source),
		GoMethod: GoMethod{
			Name:    funcDecl.Name.Name,
			Params:  buildTypeList(info, funcDecl.Type.Params, source),
			Results: buildTypeList(info, funcDecl.Type.Results, source),
			Doc:     ExtractDocs(funcDecl.Doc),
		},
	}

}

func buildReceiverList(info *types.Info, fieldList *ast.FieldList, source []byte) []string {
	receivers := []string{}

	if fieldList != nil {
		for _, t := range fieldList.List {
			receivers = append(receivers, getTypeString(t.Type, source))
		}
	}

	return receivers
}

func buildTypeList(info *types.Info, fieldList *ast.FieldList, source []byte) []*GoType {
	types := []*GoType{}

	if fieldList != nil {
		for _, t := range fieldList.List {
			goType := buildType(info, t.Type, source)

			for _, n := range getNames(t) {
				copyType := copyType(goType)
				copyType.Name = n
				types = append(types, copyType)
			}
		}
	}

	return types
}

func getNames(field *ast.Field) []string {
	if field.Names == nil || len(field.Names) == 0 {
		return []string{""}
	}

	result := []string{}
	for _, name := range field.Names {
		result = append(result, name.String())
	}

	return result
}

func getTypeString(expr ast.Expr, source []byte) string {
	return string(source[expr.Pos()-1 : expr.End()-1])
}

func getUnderlyingTypeString(info *types.Info, expr ast.Expr) string {
	if typeInfo := info.TypeOf(expr); typeInfo != nil {
		if underlying := typeInfo.Underlying(); underlying != nil {
			return underlying.String()
		}
	}

	return ""
}

func copyType(goType *GoType) *GoType {

	return &GoType{
		Type:       goType.Type,
		Inner:      goType.Inner,
		Name:       goType.Name,
		Underlying: goType.Underlying,
	}

}

func buildType(info *types.Info, expr ast.Expr, source []byte) *GoType {

	innerTypes := []*GoType{}
	typeString := getTypeString(expr, source)
	underlyingString := getUnderlyingTypeString(info, expr)

	switch specType := expr.(type) {
	case *ast.FuncType:
		innerTypes = append(innerTypes, buildTypeList(info, specType.Params, source)...)
		innerTypes = append(innerTypes, buildTypeList(info, specType.Results, source)...)
	case *ast.ArrayType:
		innerTypes = append(innerTypes, buildType(info, specType.Elt, source))
	case *ast.StructType:
		innerTypes = append(innerTypes, buildTypeList(info, specType.Fields, source)...)
	case *ast.MapType:
		innerTypes = append(innerTypes, buildType(info, specType.Key, source))
		innerTypes = append(innerTypes, buildType(info, specType.Value, source))
	case *ast.ChanType:
		innerTypes = append(innerTypes, buildType(info, specType.Value, source))
	case *ast.StarExpr:
		innerTypes = append(innerTypes, buildType(info, specType.X, source))
	case *ast.Ellipsis:
		innerTypes = append(innerTypes, buildType(info, specType.Elt, source))
	case *ast.InterfaceType:
		methods := buildMethodList(info, specType.Methods.List, source)
		for _, m := range methods {
			innerTypes = append(innerTypes, m.Params...)
			innerTypes = append(innerTypes, m.Results...)
		}

	case *ast.Ident:
	case *ast.SelectorExpr:
	default:
		fmt.Printf("Unexpected field type: `%s`,\n %#v\n", typeString, specType)
	}

	return &GoType{
		Type:       typeString,
		Underlying: underlyingString,
		Inner:      innerTypes,
	}
}

func buildGoStruct(source []byte, file *GoFile, info *types.Info, typeSpec *ast.TypeSpec, structType *ast.StructType) *GoStruct {
	goStruct := &GoStruct{
		File:   file,
		Name:   typeSpec.Name.Name,
		Fields: []*GoField{},
		Doc:    ExtractDocs(typeSpec.Doc),
	}

	// Field: A Field declaration list in a struct type, a method list in an interface type,
	// or a parameter/result declaration in a signature: https://golang.org/pkg/go/ast/#Field
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			goField := &GoField{
				Struct: goStruct,
				Name:   name.String(),
				Type:   string(source[field.Type.Pos()-1 : field.Type.End()-1]),
			}

			if field.Tag != nil {
				goTag := &GoTag{
					Field: goField,
					Value: field.Tag.Value,
				}

				goField.Tag = goTag
			}

			goStruct.Fields = append(goStruct.Fields, goField)
		}
	}

	return goStruct
}
