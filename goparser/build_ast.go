// Package goparser was taken from an open source project (https://github.com/zpatrick/go-parser) by zpatrick. Since it seemed
// that he had abandon it, I've integrated it into this project (and extended it).
package goparser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// buildVarAssignment creates GoAssignment instances from AST variable/constant declarations.
func buildVarAssignment(
	ctx *parseContext,
	file *GoFile,
	info *types.Info,
	genDecl *ast.GenDecl,
	valueSpec *ast.ValueSpec,
	src fileSource,
) []*GoAssignment {

	list := []*GoAssignment{}
	for i := range valueSpec.Names {

		goVarAssignment := &GoAssignment{
			File:     file,
			Name:     valueSpec.Names[i].Name,
			FullDecl: src.slice(genDecl.Pos(), genDecl.End()),
			Exported: isExported(valueSpec.Names[i].Name),
		}

		if specDecl := strings.TrimSpace(src.slice(valueSpec.Pos(), valueSpec.End())); specDecl != "" {
			goVarAssignment.Decl = specDecl
		} else {
			goVarAssignment.Decl = strings.TrimSpace(src.slice(genDecl.Pos(), genDecl.End()))
		}

		if genDecl.Doc != nil {
			goVarAssignment.Decl = strings.TrimSpace(src.slice(genDecl.Pos(), genDecl.End()))
			goVarAssignment.Doc = docString(ctx, genDecl.Doc, valueSpec.Pos())
		}

		if valueSpec.Doc != nil {
			goVarAssignment.Decl = strings.TrimSpace(src.slice(valueSpec.Pos(), valueSpec.End()))
			goVarAssignment.Doc = docString(ctx, valueSpec.Doc, valueSpec.Pos())
		}

		if genDecl.Tok == token.CONST {
			if decl := renderConstDecl(file, info, valueSpec, i, src); decl != "" {
				goVarAssignment.Decl = decl
			}
		}

		list = append(list, goVarAssignment)
	}

	return list
}

// renderConstDecl renders a constant declaration with its evaluated value.
func renderConstDecl(
	file *GoFile,
	info *types.Info,
	valueSpec *ast.ValueSpec,
	index int,
	src fileSource,
) string {
	if info == nil || valueSpec == nil || index < 0 || index >= len(valueSpec.Names) {
		return ""
	}

	nameIdent := valueSpec.Names[index]
	if nameIdent == nil {
		return ""
	}

	obj := info.Defs[nameIdent]
	constObj, ok := obj.(*types.Const)
	if !ok || constObj == nil {
		return ""
	}

	val := constObj.Val()
	if val == nil {
		return ""
	}

	valueText := val.String()
	if valueText == "" {
		return ""
	}

	typeText := renderConstType(file, constObj.Type())

	left := renderConstDeclLeft(valueSpec, index, typeText, src)
	if left == "" {
		left = nameIdent.Name
	}

	return fmt.Sprintf("%s = %s", left, valueText)
}

// renderConstDeclLeft renders the left-hand side of a constant declaration.
func renderConstDeclLeft(
	valueSpec *ast.ValueSpec,
	index int,
	explicitType string,
	src fileSource,
) string {
	specDecl := strings.TrimSpace(src.slice(valueSpec.Pos(), valueSpec.End()))
	if specDecl != "" && len(valueSpec.Names) == 1 {
		if idx := strings.Index(specDecl, "="); idx != -1 {
			specDecl = strings.TrimSpace(specDecl[:idx])
		}
		specDecl = strings.TrimSpace(specDecl)
		if explicitType != "" {
			return fmt.Sprintf("%s %s", valueSpec.Names[index].Name, explicitType)
		}
		return specDecl
	}

	left := valueSpec.Names[index].Name
	if valueSpec.Type != nil {
		typeText := strings.TrimSpace(src.slice(valueSpec.Type.Pos(), valueSpec.Type.End()))
		if typeText != "" {
			left = fmt.Sprintf("%s %s", left, typeText)
			explicitType = ""
		}
	}

	if explicitType != "" {
		left = fmt.Sprintf("%s %s", left, explicitType)
	}

	return left
}

// renderConstType renders the type name for a constant.
func renderConstType(file *GoFile, typ types.Type) string {
	if typ == nil {
		return ""
	}

	if basic, ok := typ.(*types.Basic); ok {
		if (basic.Info() & types.IsUntyped) != 0 {
			return ""
		}
		return basic.Name()
	}

	qualifier := func(pkg *types.Package) string {
		if pkg == nil {
			return ""
		}
		if file != nil {
			if file.Package != "" && pkg.Name() == file.Package {
				return ""
			}
			if file.FqPackage != "" && pkg.Path() == file.FqPackage {
				return ""
			}
		}
		return pkg.Name()
	}

	typeText := types.TypeString(typ, qualifier)
	if strings.HasPrefix(typeText, "untyped ") {
		return ""
	}

	return typeText
}

// buildGoImport creates a GoImport from an AST import specification.
func buildGoImport(ctx *parseContext, spec *ast.ImportSpec, file *GoFile) *GoImport {
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
		Doc:  docString(ctx, spec.Doc, spec.Pos()),
	}
}

// buildGoInterface creates a GoInterface from an AST interface type.
func buildGoInterface(
	ctx *parseContext,
	src fileSource,
	file *GoFile,
	info *types.Info,
	typeSpec *ast.TypeSpec,
	interfaceType *ast.InterfaceType,
) *GoInterface {

	methods, typeSet, typeSetDecl := buildInterfaceMembers(ctx, file, info, interfaceType.Methods, src)

	return &GoInterface{
		File:        file,
		Name:        typeSpec.Name.Name,
		Exported:    isExported(typeSpec.Name.Name),
		Methods:     methods,
		TypeParams:  buildTypeParamList(ctx, file, info, typeSpec.TypeParams, src),
		TypeSet:     typeSet,
		TypeSetDecl: typeSetDecl,
	}

}

// buildInterfaceMembers extracts methods and type sets from an interface.
func buildInterfaceMembers(
	ctx *parseContext,
	file *GoFile,
	info *types.Info,
	fieldList *ast.FieldList,
	src fileSource,
) ([]*GoMethod, []*GoType, []string) {
	methods := []*GoMethod{}
	typeSet := []*GoType{}
	typeSetDecl := []string{}

	if fieldList == nil {
		return methods, typeSet, typeSetDecl
	}

	for _, field := range fieldList.List {
		if fType, ok := field.Type.(*ast.FuncType); ok && len(field.Names) > 0 {
			name := field.Names[0].Name

			goMethod := &GoMethod{
				Name:       name,
				File:       file,
				Exported:   isExported(name),
				Params:     buildTypeList(ctx, file, info, fType.Params, src),
				Results:    buildTypeList(ctx, file, info, fType.Results, src),
				Decl:       name + src.slice(fType.Pos(), fType.End()),
				FullDecl:   name + src.slice(fType.Pos(), fType.End()),
				Doc:        docString(ctx, field.Doc, field.Pos()),
				TypeParams: buildTypeParamList(ctx, file, info, fType.TypeParams, src),
			}

			methods = append(methods, goMethod)
			continue
		}

		decl := strings.TrimSpace(src.slice(field.Type.Pos(), field.Type.End()))
		if decl != "" {
			typeSetDecl = append(typeSetDecl, decl)
		}

		for _, expr := range collectTypeSetExpr(field.Type) {
			typeSet = append(typeSet, buildType(ctx, file, info, expr, src))
		}
	}

	return methods, typeSet, typeSetDecl
}

// collectTypeSetExpr collects type expressions from a type set union.
func collectTypeSetExpr(expr ast.Expr) []ast.Expr {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.OR {
			result := collectTypeSetExpr(e.X)
			result = append(result, collectTypeSetExpr(e.Y)...)
			return result
		}
	case *ast.ParenExpr:
		return collectTypeSetExpr(e.X)
	}

	return []ast.Expr{expr}
}

// buildStructMethod creates a GoStructMethod from an AST function declaration.
func buildStructMethod(
	ctx *parseContext,
	file *GoFile,
	info *types.Info,
	funcDecl *ast.FuncDecl,
	src fileSource,
) *GoStructMethod {

	receiverTypes := buildTypeList(ctx, file, info, funcDecl.Recv, src)
	receiverStrings := make([]string, len(receiverTypes))
	for i, rt := range receiverTypes {
		receiverStrings[i] = rt.Type
	}

	return &GoStructMethod{
		Receivers:     receiverStrings,
		ReceiverTypes: receiverTypes,
		GoMethod: GoMethod{
			File:       file,
			Name:       funcDecl.Name.Name,
			Exported:   isExported(funcDecl.Name.Name),
			Params:     buildTypeList(ctx, file, info, funcDecl.Type.Params, src),
			Results:    buildTypeList(ctx, file, info, funcDecl.Type.Results, src),
			Doc:        docString(ctx, funcDecl.Doc, funcDecl.Pos()),
			TypeParams: buildTypeParamList(ctx, file, info, funcDecl.Type.TypeParams, src),
		},
	}

}

// buildTypeParamList creates a list of type parameters from a field list.
func buildTypeParamList(
	ctx *parseContext,
	file *GoFile,
	info *types.Info,
	fieldList *ast.FieldList,
	src fileSource,
) []*GoType {
	params := []*GoType{}

	if fieldList == nil {
		return params
	}

	for _, field := range fieldList.List {
		var constraint *GoType
		if field.Type != nil {
			constraint = buildType(ctx, file, info, field.Type, src)
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

// buildTypeList creates a list of GoTypes from a field list.
func buildTypeList(
	ctx *parseContext,
	file *GoFile,
	info *types.Info,
	fieldList *ast.FieldList,
	src fileSource,
) []*GoType {
	types := []*GoType{}

	if fieldList != nil {
		for _, t := range fieldList.List {
			goType := buildType(ctx, file, info, t.Type, src)

			for _, n := range getNames(t) {
				copyType := copyType(goType)
				copyType.Name = n
				types = append(types, copyType)
			}
		}
	}

	return types
}

// getNames extracts field names from an AST field.
func getNames(field *ast.Field) []string {
	if field == nil || len(field.Names) == 0 {
		return []string{""}
	}

	result := []string{}
	for _, name := range field.Names {
		result = append(result, name.String())
	}

	return result
}

// getTypeString gets the string representation of a type from source.
func getTypeString(expr ast.Expr, src fileSource) string {
	return src.slice(expr.Pos(), expr.End())
}

// getUnderlyingTypeString gets the underlying type string from type info.
func getUnderlyingTypeString(info *types.Info, expr ast.Expr) string {
	if typeInfo := info.TypeOf(expr); typeInfo != nil {
		if underlying := typeInfo.Underlying(); underlying != nil {
			return underlying.String()
		}
	}

	return ""
}

// copyType creates a shallow copy of a GoType.
func copyType(goType *GoType) *GoType {

	return &GoType{
		File:       goType.File,
		Type:       goType.Type,
		Exported:   goType.Exported,
		Inner:      goType.Inner,
		Name:       goType.Name,
		Underlying: goType.Underlying,
		Kind:       goType.Kind,
	}

}

// buildType creates a GoType from an AST expression.
func buildType(ctx *parseContext, file *GoFile, info *types.Info, expr ast.Expr, src fileSource) *GoType {

	innerTypes := []*GoType{}
	typeString := getTypeString(expr, src)
	underlyingString := getUnderlyingTypeString(info, expr)
	kind := TypeKindUnknown

	switch specType := expr.(type) {
	case *ast.FuncType:
		kind = TypeKindFunc
		innerTypes = append(innerTypes, buildTypeList(ctx, file, info, specType.Params, src)...)
		innerTypes = append(innerTypes, buildTypeList(ctx, file, info, specType.Results, src)...)
	case *ast.ArrayType:
		if specType.Len == nil {
			kind = TypeKindSlice
		} else {
			kind = TypeKindArray
		}
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Elt, src))
	case *ast.StructType:
		kind = TypeKindStruct
		innerTypes = append(innerTypes, buildTypeList(ctx, file, info, specType.Fields, src)...)
	case *ast.MapType:
		kind = TypeKindMap
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Key, src))
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Value, src))
	case *ast.ChanType:
		kind = TypeKindChan
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Value, src))
	case *ast.StarExpr:
		kind = TypeKindPointer
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.X, src))
	case *ast.Ellipsis:
		kind = TypeKindEllipsis
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Elt, src))
	case *ast.InterfaceType:
		kind = TypeKindInterface
		methods, embeds, _ := buildInterfaceMembers(ctx, file, info, specType.Methods, src)
		for _, m := range methods {
			innerTypes = append(innerTypes, m.Params...)
			innerTypes = append(innerTypes, m.Results...)
		}
		innerTypes = append(innerTypes, embeds...)
	case *ast.IndexExpr:
		kind = TypeKindIndex
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.X, src))
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Index, src))
	case *ast.IndexListExpr:
		kind = TypeKindIndexList
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.X, src))
		for _, idx := range specType.Indices {
			innerTypes = append(innerTypes, buildType(ctx, file, info, idx, src))
		}
	case *ast.UnaryExpr:
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.X, src))
	case *ast.BinaryExpr:
		kind = TypeKindBinaryExpr
		if specType.Op == token.OR {
			innerTypes = append(innerTypes, buildType(ctx, file, info, specType.X, src))
			innerTypes = append(innerTypes, buildType(ctx, file, info, specType.Y, src))
		}
	case *ast.ParenExpr:
		kind = TypeKindParen
		innerTypes = append(innerTypes, buildType(ctx, file, info, specType.X, src))
	case *ast.Ident:
		kind = TypeKindIdent
	case *ast.SelectorExpr:
		kind = TypeKindSelector
	default:
		// Log unexpected types for debugging without polluting stdout
		if file != nil && file.Module != nil {
			file.Module.AddUnresolvedDeclaration(UnresolvedDecl{
				Expr:    expr,
				Message: fmt.Sprintf("unexpected field type: %s (%T)", typeString, specType),
			})
		}
	}

	return &GoType{
		File:       file,
		Type:       typeString,
		Exported:   isExported(typeString),
		Underlying: underlyingString,
		Inner:      innerTypes,
		Kind:       kind,
	}
}

// buildGoStruct creates a GoStruct from an AST struct type.
func buildGoStruct(
	ctx *parseContext,
	src fileSource,
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
		TypeParams: buildTypeParamList(ctx, file, info, typeParams, src),
	}

	// Field: A Field declaration list in a struct type, a method list in an interface type,
	// or a parameter/result declaration in a signature: https://golang.org/pkg/go/ast/#Field
	for _, field := range structType.Fields.List {

		if len(field.Names) == 0 {
			// Derives from other struct
			typeInfo := buildType(ctx, file, info, field.Type, src)
			goField := &GoField{
				Struct:   goStruct,
				File:     file,
				Name:     "",
				Type:     src.slice(field.Type.Pos(), field.Type.End()),
				Decl:     src.slice(field.Type.Pos(), field.Type.End()),
				Doc:      docString(ctx, field.Doc, field.Pos()),
				TypeInfo: copyType(typeInfo),
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

		typeInfo := buildType(ctx, file, info, field.Type, src)

		for _, name := range field.Names {

			var anonymousStruct *GoStruct = nil
			if fld, ok := name.Obj.Decl.(*ast.Field); ok {

				if st, ok := fld.Type.(*ast.StructType); ok {
					anonymousStruct = buildGoStruct(ctx, src, file, info, name.Name, nil, st)
					anonymousStruct.Doc = docString(ctx, fld.Doc, fld.Pos())
					anonymousStruct.Decl = "struct"
				}
			}

			goField := &GoField{
				Struct:          goStruct,
				File:            file,
				Name:            name.String(),
				Exported:        isExported(name.String()),
				Type:            src.slice(field.Type.Pos(), field.Type.End()),
				Decl:            name.Name + " " + src.slice(field.Type.Pos(), field.Type.End()),
				Doc:             docString(ctx, field.Doc, field.Pos()),
				AnonymousStruct: anonymousStruct,
				TypeInfo:        copyType(typeInfo),
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
