// Package goparser was taken from an open source project (https://github.com/zpatrick/go-parser) by zpatrick. Since it seemed
// that he had abandon it, I've integrated it into this project (and extended it).
package goparser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/ioutil"
	"unicode"
)

func parseFile(
	mod *GoModule,
	path string,
	source []byte,
	file *ast.File,
	fset *token.FileSet,
	files []*ast.File,
) (*GoFile, error) {

	var err error
	if len(source) == 0 {
		source, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	// To import sources from vendor, we use "source" compile
	// https://github.com/golang/go/issues/11415#issuecomment-283445198
	//conf := types.Config{Importer: importer.For("source", nil)} // TODO: re-enable when conf.Check has been solved!
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	/* TODO: This segfaults on gopackage.go!!
	if _, err := conf.Check(file.Name.Name, fset, files, info); err != nil {
		return nil, err
	}*/

	goFile := &GoFile{
		Module:   mod,
		FilePath: path,
		Doc:      extractDocs(file.Doc),
		Decl:     "package " + file.Name.Name,
		Package:  file.Name.Name,
		Structs:  []*GoStruct{},
	}

	if mod != nil {
		goFile.FqPackage = mod.ResolvePackage(path)
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
						goStruct := buildGoStruct(source, goFile, info, typeSpec.Name.Name, typeSpec.TypeParams, structType)
						goStruct.Doc = extractDocs(declType.Doc)
						goStruct.Decl = "type " + genSpecType.Name.Name + " struct"
						goStruct.FullDecl = string(source[decl.Pos()-1 : decl.End()-1])
						goFile.Structs = append(goFile.Structs, goStruct)
					// InterfaceType: An InterfaceType node represents an interface type. https://golang.org/pkg/go/ast/#InterfaceType
					case (*ast.InterfaceType):
						interfaceType := typeSpecType
						goInterface := buildGoInterface(source, goFile, info, typeSpec, interfaceType)
						goInterface.Doc = extractDocs(declType.Doc)
						goInterface.Decl = "type " + genSpecType.Name.Name + " interface"
						goInterface.FullDecl = string(source[decl.Pos()-1 : decl.End()-1])
						goFile.Interfaces = append(goFile.Interfaces, goInterface)
						// Custom Type declaration
					case (*ast.Ident):
						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     typeSpecType.Name,
							Doc:      extractDocs(declType.Doc),
							Decl:     string(source[decl.Pos()-1 : decl.End()-1]),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, source)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					case (*ast.FuncType):
						funcType := typeSpecType

						goMethod := &GoMethod{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Decl:     string(source[decl.Pos()-1 : decl.End()-1]),
							FullDecl: string(source[decl.Pos()-1 : decl.End()-1]),
							Params:   buildTypeList(goFile, info, funcType.Params, source),
							Results:  buildTypeList(goFile, info, funcType.Results, source),
							Doc:      extractDocs(declType.Doc),
						}

						aliasParams := buildTypeParamList(goFile, info, typeSpec.TypeParams, source)
						funcParams := buildTypeParamList(goFile, info, funcType.TypeParams, source)
						goMethod.TypeParams = append(aliasParams, funcParams...)

						goFile.CustomFuncs = append(goFile.CustomFuncs, goMethod)
					case (*ast.SelectorExpr):
						selectType := typeSpecType

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     selectType.X.(*ast.Ident).Name + "." + selectType.Sel.Name,
							Doc:      extractDocs(declType.Doc),
							Decl:     string(source[decl.Pos()-1 : decl.End()-1]),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, source)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					case (*ast.ArrayType):
						//
						var length string
						if typeSpecType.Len == nil {
							length = ""
						} else {
							length = types.ExprString(typeSpecType.Len)
						}

						typeName := renderTypeName(typeSpecType.Elt)

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),

							Type: fmt.Sprintf(
								"[%s]%s", length, typeName,
							),

							Doc:  extractDocs(declType.Doc),
							Decl: string(source[decl.Pos()-1 : decl.End()-1]),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, source)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					case (*ast.MapType):

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),

							Type: fmt.Sprintf(
								"map[%s]%s",
								source[typeSpecType.Key.Pos()-1:typeSpecType.Key.End()-1],
								source[typeSpecType.Value.Pos()-1:typeSpecType.Value.End()-1],
							),

							Doc:  extractDocs(declType.Doc),
							Decl: string(source[decl.Pos()-1 : decl.End()-1]),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, source)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)

					default:

						typeExpr := types.ExprString(typeSpec.Type)

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     typeExpr,
							Doc:      extractDocs(declType.Doc),
							Decl:     string(source[decl.Pos()-1 : decl.End()-1]),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, source)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)

						if mod != nil {
							buf := bytes.NewBufferString("")

							fmt.Fprintf(
								buf,
								"not-implemented typeSpec.Type.(type) = %T, ast-dump:\n----------------------\n",
								typeSpec.Type,
							)

							ast.Fprint(buf, fset, typeSpecType, ast.NotNilFilter)
							fmt.Fprintf(buf, "----------------------")

							mod.AddUnresolvedDeclaration(UnresolvedDecl{
								Expr:    typeSpec.Type,
								Message: buf.String(),
							})
						}
					}

					// ImportSpec: An ImportSpec node represents a single package import. https://golang.org/pkg/go/ast/#ImportSpec
				case *ast.ImportSpec:
					importSpec := genSpec.(*ast.ImportSpec)
					goImport := buildGoImport(importSpec, goFile)
					goFile.ImportFullDecl = string(source[decl.Pos()-1 : decl.End()-1])
					goFile.Imports = append(goFile.Imports, goImport)
				case *ast.ValueSpec:
					valueSpec := genSpecType

					switch genDecl.Tok {
					case token.VAR:
						goFile.VarAssignments = append(goFile.VarAssignments, buildVarAssignment(goFile, genDecl, valueSpec, source)...)
					case token.CONST:

						goFile.ConstAssignments = append(goFile.ConstAssignments, buildVarAssignment(goFile, genDecl, valueSpec, source)...)
					}
				default:
					// a not-implemented genSpec.(type), ignore
				}
			}
		case *ast.FuncDecl:
			funcDecl := declType
			goStructMethod := buildStructMethod(goFile, info, funcDecl, source)
			goStructMethod.Decl = string(source[funcDecl.Type.Pos()-1 : funcDecl.Type.End()-1])
			goStructMethod.FullDecl = string(source[decl.Pos()-1 : decl.End()-1])
			goFile.StructMethods = append(goFile.StructMethods, goStructMethod)

		default:
			// a not-implemented decl.(type), ignore
		}
	}

	return goFile, nil
}

// isExported returns true if name starts with a upper-case letter
func isExported(name string) bool {

	r := rune(name[0])
	return unicode.IsUpper(r) && unicode.IsLetter(r)

}

func buildVarAssignment(
	file *GoFile,
	genDecl *ast.GenDecl,
	valueSpec *ast.ValueSpec,
	source []byte,
) []*GoAssignment {

	list := []*GoAssignment{}
	for i := range valueSpec.Names {

		goVarAssignment := &GoAssignment{
			File:     file,
			Name:     valueSpec.Names[i].Name,
			FullDecl: string(source[genDecl.Pos()-1 : genDecl.End()-1]),
			Exported: isExported(valueSpec.Names[i].Name),
		}

		if genDecl.Doc != nil {
			goVarAssignment.Decl = string(source[genDecl.Pos()-1 : genDecl.End()-1])
			goVarAssignment.Doc = extractDocs(genDecl.Doc)
		}

		if valueSpec.Doc != nil {
			goVarAssignment.Decl = string(source[valueSpec.Pos()-1 : valueSpec.End()-1])
			goVarAssignment.Doc = extractDocs(valueSpec.Doc)
		}

		list = append(list, goVarAssignment)
	}

	return list
}

func extractDocs(doc *ast.CommentGroup) string {
	d := doc.Text()
	if d == "" {
		return d
	}

	return d[:len(d)-1]
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
		Doc:  extractDocs(spec.Doc),
	}
}

func buildGoInterface(
	source []byte,
	file *GoFile,
	info *types.Info,
	typeSpec *ast.TypeSpec,
	interfaceType *ast.InterfaceType,
) *GoInterface {

	methods, typeSet := buildInterfaceMembers(file, info, interfaceType.Methods, source)

	return &GoInterface{
		File:       file,
		Name:       typeSpec.Name.Name,
		Exported:   isExported(typeSpec.Name.Name),
		Methods:    methods,
		TypeParams: buildTypeParamList(file, info, typeSpec.TypeParams, source),
		TypeSet:    typeSet,
	}

}

func buildInterfaceMembers(
	file *GoFile,
	info *types.Info,
	fieldList *ast.FieldList,
	source []byte,
) ([]*GoMethod, []*GoType) {
	methods := []*GoMethod{}
	typeSet := []*GoType{}

	if fieldList == nil {
		return methods, typeSet
	}

	for _, field := range fieldList.List {
		if fType, ok := field.Type.(*ast.FuncType); ok && len(field.Names) > 0 {
			name := field.Names[0].Name

			goMethod := &GoMethod{
				Name:       name,
				File:       file,
				Exported:   isExported(name),
				Params:     buildTypeList(file, info, fType.Params, source),
				Results:    buildTypeList(file, info, fType.Results, source),
				Decl:       name + string(source[fType.Pos()-1:fType.End()-1]),
				FullDecl:   name + string(source[fType.Pos()-1:fType.End()-1]),
				Doc:        extractDocs(field.Doc),
				TypeParams: buildTypeParamList(file, info, fType.TypeParams, source),
			}

			methods = append(methods, goMethod)
			continue
		}

		for _, expr := range flattenTypeSetExpr(field.Type) {
			typeSet = append(typeSet, buildType(file, info, expr, source))
		}
	}

	return methods, typeSet
}

func flattenTypeSetExpr(expr ast.Expr) []ast.Expr {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.OR {
			return append(flattenTypeSetExpr(e.X), flattenTypeSetExpr(e.Y)...)
		}
	case *ast.ParenExpr:
		return flattenTypeSetExpr(e.X)
	}

	return []ast.Expr{expr}
}

func buildStructMethod(
	file *GoFile,
	info *types.Info,
	funcDecl *ast.FuncDecl,
	source []byte,
) *GoStructMethod {

	return &GoStructMethod{
		Receivers: buildReceiverList(info, funcDecl.Recv, source),
		GoMethod: GoMethod{
			File:       file,
			Name:       funcDecl.Name.Name,
			Exported:   isExported(funcDecl.Name.Name),
			Params:     buildTypeList(file, info, funcDecl.Type.Params, source),
			Results:    buildTypeList(file, info, funcDecl.Type.Results, source),
			Doc:        extractDocs(funcDecl.Doc),
			TypeParams: buildTypeParamList(file, info, funcDecl.Type.TypeParams, source),
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

func buildTypeParamList(
	file *GoFile,
	info *types.Info,
	fieldList *ast.FieldList,
	source []byte,
) []*GoType {
	params := []*GoType{}

	if fieldList == nil {
		return params
	}

	for _, field := range fieldList.List {
		var constraint *GoType
		if field.Type != nil {
			constraint = buildType(file, info, field.Type, source)
		} else {
			constraint = &GoType{
				File: file,
				Type: "any",
			}
		}

		for _, name := range getNames(field) {
			if name == "" {
				continue
			}
			param := copyType(constraint)
			param.Name = name
			params = append(params, param)
		}
	}

	return params
}

func buildTypeList(
	file *GoFile,
	info *types.Info,
	fieldList *ast.FieldList,
	source []byte,
) []*GoType {
	types := []*GoType{}

	if fieldList != nil {
		for _, t := range fieldList.List {
			goType := buildType(file, info, t.Type, source)

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
		File:       goType.File,
		Type:       goType.Type,
		Exported:   goType.Exported,
		Inner:      goType.Inner,
		Name:       goType.Name,
		Underlying: goType.Underlying,
	}

}

func buildType(file *GoFile, info *types.Info, expr ast.Expr, source []byte) *GoType {

	innerTypes := []*GoType{}
	typeString := getTypeString(expr, source)
	underlyingString := getUnderlyingTypeString(info, expr)

	switch specType := expr.(type) {
	case *ast.FuncType:
		innerTypes = append(innerTypes, buildTypeList(file, info, specType.Params, source)...)
		innerTypes = append(innerTypes, buildTypeList(file, info, specType.Results, source)...)
	case *ast.ArrayType:
		innerTypes = append(innerTypes, buildType(file, info, specType.Elt, source))
	case *ast.StructType:
		innerTypes = append(innerTypes, buildTypeList(file, info, specType.Fields, source)...)
	case *ast.MapType:
		innerTypes = append(innerTypes, buildType(file, info, specType.Key, source))
		innerTypes = append(innerTypes, buildType(file, info, specType.Value, source))
	case *ast.ChanType:
		innerTypes = append(innerTypes, buildType(file, info, specType.Value, source))
	case *ast.StarExpr:
		innerTypes = append(innerTypes, buildType(file, info, specType.X, source))
	case *ast.Ellipsis:
		innerTypes = append(innerTypes, buildType(file, info, specType.Elt, source))
	case *ast.InterfaceType:
		methods, embeds := buildInterfaceMembers(file, info, specType.Methods, source)
		for _, m := range methods {
			innerTypes = append(innerTypes, m.Params...)
			innerTypes = append(innerTypes, m.Results...)
		}
		innerTypes = append(innerTypes, embeds...)
	case *ast.IndexExpr:
		innerTypes = append(innerTypes, buildType(file, info, specType.X, source))
		innerTypes = append(innerTypes, buildType(file, info, specType.Index, source))
	case *ast.IndexListExpr:
		innerTypes = append(innerTypes, buildType(file, info, specType.X, source))
		for _, idx := range specType.Indices {
			innerTypes = append(innerTypes, buildType(file, info, idx, source))
		}
	case *ast.UnaryExpr:
		innerTypes = append(innerTypes, buildType(file, info, specType.X, source))
	case *ast.BinaryExpr:
		if specType.Op == token.OR {
			innerTypes = append(innerTypes, buildType(file, info, specType.X, source))
			innerTypes = append(innerTypes, buildType(file, info, specType.Y, source))
		}
	case *ast.ParenExpr:
		innerTypes = append(innerTypes, buildType(file, info, specType.X, source))
	case *ast.Ident:
	case *ast.SelectorExpr:
	default:
		fmt.Printf("Unexpected field type: `%s`,\n %#v\n", typeString, specType)
	}

	return &GoType{
		File:       file,
		Type:       typeString,
		Exported:   isExported(typeString),
		Underlying: underlyingString,
		Inner:      innerTypes,
	}
}

func buildGoStruct(
	source []byte,
	file *GoFile,
	info *types.Info,
	structName string,
	typeParams *ast.FieldList,
	structType *ast.StructType,
) *GoStruct {
	goStruct := &GoStruct{
		File:       file,
		Name:       structName,
		Exported:   isExported(structName),
		Fields:     []*GoField{},
		TypeParams: buildTypeParamList(file, info, typeParams, source),
	}

	// Field: A Field declaration list in a struct type, a method list in an interface type,
	// or a parameter/result declaration in a signature: https://golang.org/pkg/go/ast/#Field
	for _, field := range structType.Fields.List {

		if len(field.Names) == 0 {
			// Derives from other struct
			goField := &GoField{
				Struct: goStruct,
				File:   file,
				Name:   "",
				Type:   string(source[field.Type.Pos()-1 : field.Type.End()-1]),
				Decl:   string(source[field.Type.Pos()-1 : field.Type.End()-1]),
				Doc:    extractDocs(field.Doc),
			}

			goField.Exported = isExported(goField.Type)

			if field.Tag != nil {
				goTag := &GoTag{
					File:  file,
					Field: goField,
					Value: field.Tag.Value,
				}

				goField.Tag = goTag
			}
			goStruct.Fields = append(goStruct.Fields, goField)
		}

		for _, name := range field.Names {

			var nested *GoStruct = nil
			if fld, ok := name.Obj.Decl.(*ast.Field); ok {

				if st, ok := fld.Type.(*ast.StructType); ok {
					nested = buildGoStruct(source, file, info, name.Name, nil, st)
					nested.Doc = extractDocs(fld.Doc)
					nested.Decl = "struct"
				}
			}

			goField := &GoField{
				Struct:   goStruct,
				File:     file,
				Name:     name.String(),
				Exported: isExported(name.String()),
				Type:     string(source[field.Type.Pos()-1 : field.Type.End()-1]),
				Decl:     name.Name + " " + string(source[field.Type.Pos()-1:field.Type.End()-1]),
				Doc:      extractDocs(field.Doc),
				Nested:   nested,
			}

			if field.Tag != nil {
				goTag := &GoTag{
					File:  file,
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
