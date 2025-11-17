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

// parseContext holds parsing state that was previously global.
// It is passed through the call chain and is safe for concurrent use
// (each goroutine gets its own context).
type parseContext struct {
	// docMode is the documentation concatenation mode
	docMode DocConcatenationMode
	// docCtx is the current documentation context for this parse operation
	docCtx *docConcatContext
}

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

// parseFileWithContext parses a Go source file with an explicit parse context.
// This is thread-safe and should be used by all new code.
func parseFileWithContext(
	ctx *parseContext,
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

	// Set up documentation context for this parse operation
	docCtx := &docConcatContext{
		mode:     ctx.docMode,
		comments: file.Comments,
		src:      src,
	}
	ctx.docCtx = docCtx

	goFile := &GoFile{
		Module:    mod,
		FilePath:  path,
		Doc:       docString(ctx, file.Doc, file.Package),
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
						goStruct := buildGoStruct(ctx, src, goFile, info, typeSpec.Name.Name, typeSpec.TypeParams, structType)
						goStruct.Doc = docString(ctx, declType.Doc, decl.Pos())
						goStruct.Decl = "type " + NameWithTypeParams(genSpecType.Name.Name, goStruct.TypeParams) + " struct"
						goStruct.FullDecl = src.slice(decl.Pos(), decl.End())
						goFile.Structs = append(goFile.Structs, goStruct)
					// InterfaceType: An InterfaceType node represents an interface type. https://golang.org/pkg/go/ast/#InterfaceType
					case (*ast.InterfaceType):
						interfaceType := typeSpecType
						goInterface := buildGoInterface(ctx, src, goFile, info, typeSpec, interfaceType)
						goInterface.Doc = docString(ctx, declType.Doc, decl.Pos())
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
							Doc:      docString(ctx, declType.Doc, decl.Pos()),
							Decl:     src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(ctx, goFile, info, typeSpec.TypeParams, src)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)
					case (*ast.FuncType):
						funcType := typeSpecType

						goMethod := &GoMethod{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Decl:     src.slice(decl.Pos(), decl.End()),
							FullDecl: src.slice(decl.Pos(), decl.End()),
							Params:   buildTypeList(ctx, goFile, info, funcType.Params, src),
							Results:  buildTypeList(ctx, goFile, info, funcType.Results, src),
							Doc:      docString(ctx, declType.Doc, decl.Pos()),
						}

						aliasParams := buildTypeParamList(ctx, goFile, info, typeSpec.TypeParams, src)
						funcParams := buildTypeParamList(ctx, goFile, info, funcType.TypeParams, src)
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
							Doc:      docString(ctx, declType.Doc, decl.Pos()),
							Decl:     src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(ctx, goFile, info, typeSpec.TypeParams, src)

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

							Doc:  docString(ctx, declType.Doc, decl.Pos()),
							Decl: src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(ctx, goFile, info, typeSpec.TypeParams, src)

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

							Doc:  docString(ctx, declType.Doc, decl.Pos()),
							Decl: src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(ctx, goFile, info, typeSpec.TypeParams, src)

						goFile.CustomTypes = append(goFile.CustomTypes, goCustomType)

					default:

						typeExpr := types.ExprString(typeSpec.Type)

						goCustomType := &GoCustomType{
							File:     goFile,
							Name:     genSpecType.Name.Name,
							Exported: isExported(genSpecType.Name.Name),
							Type:     typeExpr,
							Doc:      docString(ctx, declType.Doc, decl.Pos()),
							Decl:     src.slice(decl.Pos(), decl.End()),
						}

						goCustomType.TypeParams = buildTypeParamList(ctx, goFile, info, typeSpec.TypeParams, src)

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
					goImport := buildGoImport(ctx, importSpec, goFile)
					goFile.ImportFullDecl = src.slice(decl.Pos(), decl.End())
					goFile.Imports = append(goFile.Imports, goImport)
				case *ast.ValueSpec:
					valueSpec := genSpecType

					switch genDecl.Tok {
					case token.VAR:
						goFile.VarAssignments = append(goFile.VarAssignments, buildVarAssignment(ctx, goFile, info, genDecl, valueSpec, src)...)
					case token.CONST:

						goFile.ConstAssignments = append(goFile.ConstAssignments, buildVarAssignment(ctx, goFile, info, genDecl, valueSpec, src)...)
					}
				default:
					// a not-implemented genSpec.(type), ignore
				}
			}
		case *ast.FuncDecl:
			funcDecl := declType
			goStructMethod := buildStructMethod(ctx, goFile, info, funcDecl, src)
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
