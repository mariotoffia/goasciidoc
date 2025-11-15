// Package goparser was taken from an open source project (https://github.com/zpatrick/go-parser) by zpatrick. Since it seemed
// that he had abandon it, I've integrated it into this project (and extended it).
package goparser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"runtime"
	"strings"
	"unicode"
)

func goVersionForTypes(mod *GoModule) string {
	if mod != nil && strings.TrimSpace(mod.GoVersion) != "" {
		if v := canonicalGoVersion(mod.GoVersion); v != "" {
			return v
		}
	}

	version := runtime.Version()
	if !strings.HasPrefix(version, "go") {
		return ""
	}

	return canonicalGoVersion(strings.TrimPrefix(version, "go"))
}

func canonicalGoVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "go")
	if version == "" {
		return ""
	}

	stop := len(version)
	for i := 0; i < len(version); i++ {
		if version[i] == '.' {
			continue
		}
		if version[i] < '0' || version[i] > '9' {
			stop = i
			break
		}
	}
	version = version[:stop]

	parts := strings.Split(version, ".")
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return parts[0] + "." + parts[1]
	}
}

func typeCheckPackage(
	mod *GoModule,
	fset *token.FileSet,
	files []*ast.File,
	debug DebugFunc,
) (*types.Info, error) {
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	if len(files) == 0 {
		return info, nil
	}

	conf := types.Config{
		FakeImportC:              true,
		DisableUnusedImportCheck: true,
	}

	conf.Importer = getSharedModuleImporter(mod, debug)

	if debug != nil {
		conf.Error = func(err error) {
			if err != nil {
				debugf(debug, "typeCheck: diagnostic %v", err)
			}
		}
	}

	if goVersion := goVersionForTypes(mod); goVersion != "" {
		conf.GoVersion = goVersion
	}

	pkgName := files[0].Name.Name
	debugf(debug, "typeCheck: start %s (%d file(s))", pkgName, len(files))
	_, err := conf.Check(pkgName, fset, files, info)
	if err != nil {
		debugf(debug, "typeCheck: completed %s with error: %v", pkgName, err)
	} else {
		debugf(debug, "typeCheck: completed %s", pkgName)
	}
	return info, err
}

type fileSource struct {
	data []byte
	fset *token.FileSet
}

type docConcatContext struct {
	mode     DocConcatenationMode
	comments []*ast.CommentGroup
	src      fileSource
}

var currentDocContext docConcatContext

func (fs fileSource) slice(start, end token.Pos) string {
	if len(fs.data) == 0 || fs.fset == nil || !start.IsValid() || !end.IsValid() {
		return ""
	}

	startOffset := fs.fset.PositionFor(start, false).Offset
	endOffset := fs.fset.PositionFor(end, false).Offset
	if startOffset < 0 || endOffset < startOffset {
		return ""
	}

	if startOffset > len(fs.data) {
		return ""
	}

	if endOffset > len(fs.data) {
		endOffset = len(fs.data)
	}

	return string(fs.data[startOffset:endOffset])
}

func parseFile(
	mod *GoModule,
	path string,
	source []byte,
	file *ast.File,
	fset *token.FileSet,
	info *types.Info,
) (*GoFile, error) {

	var err error
	if len(source) == 0 {
		source, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	if info == nil {
		info = &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}
	}

	src := fileSource{
		data: source,
		fset: fset,
	}

	prevCtx := currentDocContext
	currentDocContext = docConcatContext{
		mode:     activeDocConcatenation,
		comments: file.Comments,
		src:      src,
	}
	defer func() { currentDocContext = prevCtx }()

	goFile := &GoFile{
		Module:    mod,
		FilePath:  path,
		Doc:       docString(file.Doc, file.Package),
		Decl:      "package " + file.Name.Name,
		Package:   file.Name.Name,
		BuildTags: extractBuildTagsFromComments(file.Comments),
		Structs:   []*GoStruct{},
	}

	if mod != nil {
		if fq, err := mod.ResolvePackage(path); err == nil {
			goFile.FqPackage = fq
		} else {
			mod.AddUnresolvedDeclaration(UnresolvedDecl{
				Message: fmt.Sprintf("resolve package: %v", err),
			})
		}
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
						goStruct := buildGoStruct(src, goFile, info, typeSpec.Name.Name, typeSpec.TypeParams, structType)
						goStruct.Doc = docString(declType.Doc, decl.Pos())
						goStruct.Decl = "type " + NameWithTypeParams(genSpecType.Name.Name, goStruct.TypeParams) + " struct"
						goStruct.FullDecl = src.slice(decl.Pos(), decl.End())
						goFile.Structs = append(goFile.Structs, goStruct)
					// InterfaceType: An InterfaceType node represents an interface type. https://golang.org/pkg/go/ast/#InterfaceType
					case (*ast.InterfaceType):
						interfaceType := typeSpecType
						goInterface := buildGoInterface(src, goFile, info, typeSpec, interfaceType)
						goInterface.Doc = docString(declType.Doc, decl.Pos())
						goInterface.Decl = "type " + NameWithTypeParams(genSpecType.Name.Name, goInterface.TypeParams) + " interface"
						goInterface.FullDecl = src.slice(decl.Pos(), decl.End())
						goFile.Interfaces = append(goFile.Interfaces, goInterface)
						// Custom Type declaration
					case (*ast.Ident):
						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     typeSpecType.Name,
							Doc:      docString(declType.Doc, decl.Pos()),
							Decl:     src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, src)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					case (*ast.FuncType):
						funcType := typeSpecType

						goMethod := &GoMethod{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Decl:     src.slice(decl.Pos(), decl.End()),
							FullDecl: src.slice(decl.Pos(), decl.End()),
							Params:   buildTypeList(goFile, info, funcType.Params, src),
							Results:  buildTypeList(goFile, info, funcType.Results, src),
							Doc:      docString(declType.Doc, decl.Pos()),
						}

						aliasParams := buildTypeParamList(goFile, info, typeSpec.TypeParams, src)
						funcParams := buildTypeParamList(goFile, info, funcType.TypeParams, src)
						goMethod.TypeParams = append(aliasParams, funcParams...)

						goFile.CustomFuncs = append(goFile.CustomFuncs, goMethod)
					case (*ast.SelectorExpr):
						selectType := typeSpecType

						var typeName string
						if ident, ok := selectType.X.(*ast.Ident); ok {
							typeName = ident.Name + "." + selectType.Sel.Name
						} else {
							// Fallback to ExprString for complex selector expressions
							typeName = types.ExprString(selectType)
						}

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     typeName,
							Doc:      docString(declType.Doc, decl.Pos()),
							Decl:     src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, src)

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

							Doc:  docString(declType.Doc, decl.Pos()),
							Decl: src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, src)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					case (*ast.MapType):

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),

							Type: fmt.Sprintf(
								"map[%s]%s",
								src.slice(typeSpecType.Key.Pos(), typeSpecType.Key.End()),
								src.slice(typeSpecType.Value.Pos(), typeSpecType.Value.End()),
							),

							Doc:  docString(declType.Doc, decl.Pos()),
							Decl: src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, src)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)

					default:

						typeExpr := types.ExprString(typeSpec.Type)

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     typeExpr,
							Doc:      docString(declType.Doc, decl.Pos()),
							Decl:     src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(goFile, info, typeSpec.TypeParams, src)

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
					goFile.ImportFullDecl = src.slice(decl.Pos(), decl.End())
					goFile.Imports = append(goFile.Imports, goImport)
				case *ast.ValueSpec:
					valueSpec := genSpecType

					switch genDecl.Tok {
					case token.VAR:
						goFile.VarAssignments = append(goFile.VarAssignments, buildVarAssignment(goFile, info, genDecl, valueSpec, src)...)
					case token.CONST:

						goFile.ConstAssignments = append(goFile.ConstAssignments, buildVarAssignment(goFile, info, genDecl, valueSpec, src)...)
					}
				default:
					// a not-implemented genSpec.(type), ignore
				}
			}
		case *ast.FuncDecl:
			funcDecl := declType
			goStructMethod := buildStructMethod(goFile, info, funcDecl, src)
			goStructMethod.Decl = src.slice(funcDecl.Type.Pos(), funcDecl.Type.End())
			goStructMethod.FullDecl = src.slice(decl.Pos(), decl.End())
			goFile.StructMethods = append(goFile.StructMethods, goStructMethod)

		default:
			// a not-implemented decl.(type), ignore
		}
	}

	return goFile, nil
}

// isExported returns true if name starts with a upper-case letter
func isExported(name string) bool {
	identifier := normalizeExportName(name)
	for _, r := range identifier {
		return unicode.IsLetter(r) && unicode.IsUpper(r)
	}
	return false
}

func normalizeExportName(name string) string {
	s := strings.TrimSpace(name)
	for {
		original := s
		switch {
		case strings.HasPrefix(s, "..."):
			s = strings.TrimSpace(s[3:])
			continue
		case strings.HasPrefix(s, "*"):
			s = strings.TrimSpace(s[1:])
			continue
		case strings.HasPrefix(s, "&"):
			s = strings.TrimSpace(s[1:])
			continue
		case strings.HasPrefix(s, "[]"):
			s = strings.TrimSpace(s[2:])
			continue
		case strings.HasPrefix(s, "<-chan"):
			s = strings.TrimSpace(s[len("<-chan"):])
			continue
		case strings.HasPrefix(s, "chan<-"):
			s = strings.TrimSpace(s[len("chan<-"):])
			continue
		case strings.HasPrefix(s, "chan "):
			s = strings.TrimSpace(s[len("chan "):])
			continue
		case strings.HasPrefix(s, "chan"):
			s = strings.TrimSpace(s[len("chan"):])
			continue
		case strings.HasPrefix(s, "map["):
			if trimmed, ok := trimMapPrefix(s); ok {
				s = trimmed
				continue
			}
		case strings.HasPrefix(s, "("):
			if idx := matchingDelimiterIndex(s, '(', ')'); idx != -1 {
				s = strings.TrimSpace(s[1:idx])
				continue
			}
		}
		if s == original {
			break
		}
	}

	if idx := strings.Index(s, "["); idx != -1 {
		s = s[:idx]
	}

	if idx := strings.LastIndex(s, "."); idx != -1 {
		s = s[idx+1:]
	}

	return strings.TrimSpace(s)
}

func trimMapPrefix(s string) (string, bool) {
	if !strings.HasPrefix(s, "map[") {
		return s, false
	}
	depth := 1
	for i := len("map["); i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return strings.TrimSpace(s[i+1:]), true
			}
		}
	}
	return s, false
}

func matchingDelimiterIndex(s string, open, close rune) int {
	depth := 0
	for i, r := range s {
		switch r {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func buildVarAssignment(
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
			goVarAssignment.Doc = docString(genDecl.Doc, valueSpec.Pos())
		}

		if valueSpec.Doc != nil {
			goVarAssignment.Decl = strings.TrimSpace(src.slice(valueSpec.Pos(), valueSpec.End()))
			goVarAssignment.Doc = docString(valueSpec.Doc, valueSpec.Pos())
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

func extractDocs(doc *ast.CommentGroup) string {
	d := doc.Text()
	if d == "" {
		return d
	}

	return d[:len(d)-1]
}

func docString(doc *ast.CommentGroup, declPos token.Pos) string {
	if doc == nil {
		return ""
	}
	base := extractDocs(doc)
	if currentDocContext.mode != DocConcatenationFull {
		return base
	}
	comments := currentDocContext.comments
	if len(comments) == 0 {
		return base
	}
	index := -1
	for i, group := range comments {
		if group == doc {
			index = i
			break
		}
	}
	if index == -1 {
		return base
	}

	type segment struct {
		text    string
		between string
	}

	var builder strings.Builder

	cursorStart := doc.Pos()
	preceding := []segment{}
	for i := index - 1; i >= 0; i-- {
		group := comments[i]
		if group == nil || !group.End().IsValid() {
			continue
		}
		between := currentDocContext.src.slice(group.End(), cursorStart)
		if strings.TrimSpace(between) != "" {
			break
		}
		text := extractDocs(group)
		if text == "" {
			cursorStart = group.Pos()
			continue
		}
		preceding = append(preceding, segment{text: text, between: between})
		cursorStart = group.Pos()
	}
	for i := len(preceding) - 1; i >= 0; i-- {
		builder.WriteString(preceding[i].text)
		builder.WriteString(preceding[i].between)
	}

	builder.WriteString(base)
	cursor := doc.End()

	for _, group := range comments[index+1:] {
		if group == nil || !group.Pos().IsValid() {
			continue
		}
		if declPos.IsValid() && group.Pos() >= declPos {
			break
		}
		between := currentDocContext.src.slice(cursor, group.Pos())
		if strings.TrimSpace(between) != "" {
			break
		}
		text := extractDocs(group)
		if text == "" {
			cursor = group.End()
			continue
		}
		builder.WriteString(between)
		builder.WriteString(text)
		cursor = group.End()
	}

	return builder.String()
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
		Doc:  docString(spec.Doc, spec.Pos()),
	}
}

func buildGoInterface(
	src fileSource,
	file *GoFile,
	info *types.Info,
	typeSpec *ast.TypeSpec,
	interfaceType *ast.InterfaceType,
) *GoInterface {

	methods, typeSet, typeSetDecl := buildInterfaceMembers(file, info, interfaceType.Methods, src)

	return &GoInterface{
		File:        file,
		Name:        typeSpec.Name.Name,
		Exported:    isExported(typeSpec.Name.Name),
		Methods:     methods,
		TypeParams:  buildTypeParamList(file, info, typeSpec.TypeParams, src),
		TypeSet:     typeSet,
		TypeSetDecl: typeSetDecl,
	}

}

func buildInterfaceMembers(
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
				Params:     buildTypeList(file, info, fType.Params, src),
				Results:    buildTypeList(file, info, fType.Results, src),
				Decl:       name + src.slice(fType.Pos(), fType.End()),
				FullDecl:   name + src.slice(fType.Pos(), fType.End()),
				Doc:        docString(field.Doc, field.Pos()),
				TypeParams: buildTypeParamList(file, info, fType.TypeParams, src),
			}

			methods = append(methods, goMethod)
			continue
		}

		decl := strings.TrimSpace(src.slice(field.Type.Pos(), field.Type.End()))
		if decl != "" {
			typeSetDecl = append(typeSetDecl, decl)
		}

		for _, expr := range collectTypeSetExpr(field.Type) {
			typeSet = append(typeSet, buildType(file, info, expr, src))
		}
	}

	return methods, typeSet, typeSetDecl
}

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

func buildStructMethod(
	file *GoFile,
	info *types.Info,
	funcDecl *ast.FuncDecl,
	src fileSource,
) *GoStructMethod {

	receiverTypes := buildTypeList(file, info, funcDecl.Recv, src)
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
			Params:     buildTypeList(file, info, funcDecl.Type.Params, src),
			Results:    buildTypeList(file, info, funcDecl.Type.Results, src),
			Doc:        docString(funcDecl.Doc, funcDecl.Pos()),
			TypeParams: buildTypeParamList(file, info, funcDecl.Type.TypeParams, src),
		},
	}

}

func buildTypeParamList(
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
			constraint = buildType(file, info, field.Type, src)
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
	src fileSource,
) []*GoType {
	types := []*GoType{}

	if fieldList != nil {
		for _, t := range fieldList.List {
			goType := buildType(file, info, t.Type, src)

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
	if field == nil || len(field.Names) == 0 {
		return []string{""}
	}

	result := []string{}
	for _, name := range field.Names {
		result = append(result, name.String())
	}

	return result
}

func getTypeString(expr ast.Expr, src fileSource) string {
	return src.slice(expr.Pos(), expr.End())
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
		Kind:       goType.Kind,
	}

}

func buildType(file *GoFile, info *types.Info, expr ast.Expr, src fileSource) *GoType {

	innerTypes := []*GoType{}
	typeString := getTypeString(expr, src)
	underlyingString := getUnderlyingTypeString(info, expr)
	kind := TypeKindUnknown

	switch specType := expr.(type) {
	case *ast.FuncType:
		kind = TypeKindFunc
		innerTypes = append(innerTypes, buildTypeList(file, info, specType.Params, src)...)
		innerTypes = append(innerTypes, buildTypeList(file, info, specType.Results, src)...)
	case *ast.ArrayType:
		if specType.Len == nil {
			kind = TypeKindSlice
		} else {
			kind = TypeKindArray
		}
		innerTypes = append(innerTypes, buildType(file, info, specType.Elt, src))
	case *ast.StructType:
		kind = TypeKindStruct
		innerTypes = append(innerTypes, buildTypeList(file, info, specType.Fields, src)...)
	case *ast.MapType:
		kind = TypeKindMap
		innerTypes = append(innerTypes, buildType(file, info, specType.Key, src))
		innerTypes = append(innerTypes, buildType(file, info, specType.Value, src))
	case *ast.ChanType:
		kind = TypeKindChan
		innerTypes = append(innerTypes, buildType(file, info, specType.Value, src))
	case *ast.StarExpr:
		kind = TypeKindPointer
		innerTypes = append(innerTypes, buildType(file, info, specType.X, src))
	case *ast.Ellipsis:
		kind = TypeKindEllipsis
		innerTypes = append(innerTypes, buildType(file, info, specType.Elt, src))
	case *ast.InterfaceType:
		kind = TypeKindInterface
		methods, embeds, _ := buildInterfaceMembers(file, info, specType.Methods, src)
		for _, m := range methods {
			innerTypes = append(innerTypes, m.Params...)
			innerTypes = append(innerTypes, m.Results...)
		}
		innerTypes = append(innerTypes, embeds...)
	case *ast.IndexExpr:
		kind = TypeKindIndex
		innerTypes = append(innerTypes, buildType(file, info, specType.X, src))
		innerTypes = append(innerTypes, buildType(file, info, specType.Index, src))
	case *ast.IndexListExpr:
		kind = TypeKindIndexList
		innerTypes = append(innerTypes, buildType(file, info, specType.X, src))
		for _, idx := range specType.Indices {
			innerTypes = append(innerTypes, buildType(file, info, idx, src))
		}
	case *ast.UnaryExpr:
		innerTypes = append(innerTypes, buildType(file, info, specType.X, src))
	case *ast.BinaryExpr:
		kind = TypeKindBinaryExpr
		if specType.Op == token.OR {
			innerTypes = append(innerTypes, buildType(file, info, specType.X, src))
			innerTypes = append(innerTypes, buildType(file, info, specType.Y, src))
		}
	case *ast.ParenExpr:
		kind = TypeKindParen
		innerTypes = append(innerTypes, buildType(file, info, specType.X, src))
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

func buildGoStruct(
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
		TypeParams: buildTypeParamList(file, info, typeParams, src),
	}

	// Field: A Field declaration list in a struct type, a method list in an interface type,
	// or a parameter/result declaration in a signature: https://golang.org/pkg/go/ast/#Field
	for _, field := range structType.Fields.List {

		if len(field.Names) == 0 {
			// Derives from other struct
			typeInfo := buildType(file, info, field.Type, src)
			goField := &GoField{
				Struct:   goStruct,
				File:     file,
				Name:     "",
				Type:     src.slice(field.Type.Pos(), field.Type.End()),
				Decl:     src.slice(field.Type.Pos(), field.Type.End()),
				Doc:      docString(field.Doc, field.Pos()),
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

		typeInfo := buildType(file, info, field.Type, src)

		for _, name := range field.Names {

			var anonymousStruct *GoStruct = nil
			if fld, ok := name.Obj.Decl.(*ast.Field); ok {

				if st, ok := fld.Type.(*ast.StructType); ok {
					anonymousStruct = buildGoStruct(src, file, info, name.Name, nil, st)
					anonymousStruct.Doc = docString(fld.Doc, fld.Pos())
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
				Doc:             docString(field.Doc, field.Pos()),
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

// extractBuildTagsFromComments extracts build tags from file comments
func extractBuildTagsFromComments(comments []*ast.CommentGroup) []string {
	if len(comments) == 0 {
		return nil
	}

	var tags []string
	seen := make(map[string]bool)

	for _, cg := range comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(c.Text)

			// Check for //go:build directive (newer style)
			if strings.HasPrefix(text, "//go:build ") {
				constraint := strings.TrimPrefix(text, "//go:build ")
				constraint = strings.TrimSpace(constraint)
				if constraint != "" && !seen[constraint] {
					tags = append(tags, constraint)
					seen[constraint] = true
				}
			}

			// Check for // +build directive (older style)
			if strings.HasPrefix(text, "// +build ") {
				constraint := strings.TrimPrefix(text, "// +build ")
				constraint = strings.TrimSpace(constraint)
				if constraint != "" && !seen[constraint] {
					tags = append(tags, constraint)
					seen[constraint] = true
				}
			}
		}
	}

	return tags
}
