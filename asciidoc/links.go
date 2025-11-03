package asciidoc

import (
	"fmt"
	"html"
	htmltemplate "html/template"
	"path"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

type TypeLinkMode int

const (
	TypeLinksDisabled TypeLinkMode = iota
	TypeLinksInternal
	TypeLinksInternalExternal
)

var builtinTypes = map[string]struct{}{
	"bool":       {},
	"byte":       {},
	"complex64":  {},
	"complex128": {},
	"error":      {},
	"float32":    {},
	"float64":    {},
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"rune":       {},
	"string":     {},
	"uint":       {},
	"uint8":      {},
	"uint16":     {},
	"uint32":     {},
	"uint64":     {},
	"uintptr":    {},
	"any":        {},
}

type SignatureKind int

const (
	SignatureKindUnknown SignatureKind = iota
	SignatureKindFunction
	SignatureKindMethod
	SignatureKindFuncType
)

type SignatureDoc struct {
	Raw      string
	Kind     SignatureKind
	Segments []SignatureSegment
}

type SignatureSegmentKind int

const (
	SegmentKindText SignatureSegmentKind = iota
	SegmentKindKeyword
	SegmentKindName
	SegmentKindParams
	SegmentKindReceiver
	SegmentKindResult
	SegmentKindTypeName
)

type SignatureSegment struct {
	Kind    SignatureSegmentKind
	Content htmltemplate.HTML
}

type SignatureHighlightToken struct {
	Class   string
	Content htmltemplate.HTML
}

type SignatureHighlightBlock struct {
	WrapperClass string
	Tokens       []SignatureHighlightToken
}

func baseTypeIdentifier(expr string) string {
	s := strings.TrimSpace(expr)
	for {
		switch {
		case strings.HasPrefix(s, "..."):
			s = strings.TrimSpace(s[3:])
		case strings.HasPrefix(s, "*"):
			s = strings.TrimSpace(s[1:])
		case strings.HasPrefix(s, "[]"):
			s = strings.TrimSpace(s[2:])
		default:
			goto afterPrefixes
		}
	}
afterPrefixes:
	if idx := strings.Index(s, "["); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.LastIndex(s, "."); idx != -1 {
		s = s[idx+1:]
	}
	return strings.TrimSpace(s)
}

func (t *TemplateContext) typeAnchorFor(name string, file *goparser.GoFile) string {
	pkgPath := t.packagePathForFile(file)
	if pkgPath == "" {
		return ""
	}
	return fmt.Sprintf("[[%s]]\n", anchorID(pkgPath, name))
}

func (t *TemplateContext) packagePathForFile(file *goparser.GoFile) string {
	if file == nil {
		file = t.File
	}
	if file == nil {
		return ""
	}
	if file.FqPackage != "" {
		return file.FqPackage
	}
	if file.Module != nil {
		if pkg, err := file.Module.ResolvePackage(file.FilePath); err == nil {
			return pkg
		}
	}
	if t.Module != nil {
		if pkg, err := t.Module.ResolvePackage(file.FilePath); err == nil {
			return pkg
		}
	}
	if file.Package != "" {
		return file.Package
	}
	if t.Package != nil && t.Package.GoFile.FqPackage != "" {
		return t.Package.GoFile.FqPackage
	}
	if t.File != nil && t.File.FqPackage != "" {
		return t.File.FqPackage
	}
	return ""
}

func anchorID(pkgPath, typeName string) string {
	identifier := pkgPath
	if identifier != "" {
		identifier += "."
	}
	identifier += typeName

	var b strings.Builder
	b.Grow(len(identifier))
	for _, r := range identifier {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

func (t *TemplateContext) importMap(file *goparser.GoFile) map[string]string {
	if file == nil {
		file = t.File
	}
	if file == nil {
		return map[string]string{}
	}
	if t.importCache == nil {
		t.importCache = map[*goparser.GoFile]map[string]string{}
	}
	if mp, ok := t.importCache[file]; ok {
		return mp
	}
	mp := map[string]string{}
	for _, imp := range file.Imports {
		if imp == nil || imp.Path == "" {
			continue
		}
		if imp.Name == "." || imp.Name == "_" {
			continue
		}
		if imp.Name != "" {
			mp[imp.Name] = imp.Path
		}
		base := path.Base(imp.Path)
		if base != "" {
			mp[base] = imp.Path
		}
	}
	t.importCache[file] = mp
	return mp
}

func (t *TemplateContext) importPathForAlias(alias string, file *goparser.GoFile) string {
	if alias == "" {
		return ""
	}
	return t.importMap(file)[alias]
}

func (t *TemplateContext) isInternalImport(path string) bool {
	if path == "" {
		return false
	}
	if t.Module != nil && strings.HasPrefix(path, t.Module.Name) {
		return true
	}
	if t.File != nil && t.File.Module != nil && strings.HasPrefix(path, t.File.Module.Name) {
		return true
	}
	return false
}

func (t *TemplateContext) typeParamSet(lists ...[]*goparser.GoType) map[string]struct{} {
	set := map[string]struct{}{}
	for _, list := range lists {
		for _, param := range list {
			if param == nil {
				continue
			}
			name := strings.TrimSpace(param.Name)
			if name == "" {
				continue
			}
			set[name] = struct{}{}
		}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}

func (t *TemplateContext) renderTypeWithScope(gt *goparser.GoType, scope map[string]struct{}) string {
	if gt == nil {
		return ""
	}
	if t.Config == nil || t.Config.TypeLinks == TypeLinksDisabled {
		return gt.Type
	}
	return t.renderType(gt, scope)
}

func (t *TemplateContext) renderType(gt *goparser.GoType, scope map[string]struct{}) string {
	if gt == nil {
		return ""
	}
	switch gt.Kind {
	case goparser.TypeKindPointer:
		if len(gt.Inner) > 0 {
			return "*" + t.renderType(gt.Inner[0], scope)
		}
	case goparser.TypeKindSlice:
		if len(gt.Inner) > 0 {
			return "[]" + t.renderType(gt.Inner[0], scope)
		}
	case goparser.TypeKindArray:
		if len(gt.Inner) > 0 {
			if idx := strings.Index(gt.Type, "]"); idx != -1 {
				prefix := gt.Type[:idx+1]
				return prefix + t.renderType(gt.Inner[0], scope)
			}
			return gt.Type
		}
	case goparser.TypeKindMap:
		if len(gt.Inner) >= 2 {
			return "map[" + t.renderType(gt.Inner[0], scope) + "]" + t.renderType(gt.Inner[1], scope)
		}
	case goparser.TypeKindChan:
		if len(gt.Inner) > 0 {
			switch {
			case strings.HasPrefix(gt.Type, "<-chan "):
				return "<-chan " + t.renderType(gt.Inner[0], scope)
			case strings.HasPrefix(gt.Type, "chan<- "):
				return "chan<- " + t.renderType(gt.Inner[0], scope)
			case strings.HasPrefix(gt.Type, "chan "):
				return "chan " + t.renderType(gt.Inner[0], scope)
			default:
				return "chan " + t.renderType(gt.Inner[0], scope)
			}
		}
	case goparser.TypeKindEllipsis:
		if len(gt.Inner) > 0 {
			return "..." + t.renderType(gt.Inner[0], scope)
		}
	case goparser.TypeKindIndex:
		if len(gt.Inner) >= 2 {
			return t.renderType(gt.Inner[0], scope) + "[" + t.renderType(gt.Inner[1], scope) + "]"
		}
	case goparser.TypeKindIndexList:
		if len(gt.Inner) > 0 {
			parts := make([]string, 0, len(gt.Inner)-1)
			for _, inner := range gt.Inner[1:] {
				parts = append(parts, t.renderType(inner, scope))
			}
			return t.renderType(gt.Inner[0], scope) + "[" + strings.Join(parts, ", ") + "]"
		}
	case goparser.TypeKindIdent, goparser.TypeKindSelector:
		return t.linkIdentifier(gt.Type, gt.File, scope)
	default:
		if len(gt.Inner) == 0 {
			return t.linkIdentifier(gt.Type, gt.File, scope)
		}
	}
	return gt.Type
}

func (t *TemplateContext) linkIdentifier(name string, file *goparser.GoFile, scope map[string]struct{}) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return name
	}
	prefix := ""
	if strings.HasPrefix(trimmed, "~") {
		prefix = "~"
		trimmed = strings.TrimSpace(trimmed[1:])
	}
	if scope != nil {
		if _, ok := scope[trimmed]; ok {
			return prefix + trimmed
		}
	}
	if _, ok := builtinTypes[trimmed]; ok {
		return prefix + trimmed
	}

	alias := ""
	typeName := trimmed
	if idx := strings.Index(trimmed, "."); idx != -1 {
		alias = trimmed[:idx]
		typeName = trimmed[idx+1:]
	}

	if alias == "" {
		pkgPath := t.packagePathForFile(file)
		if pkgPath == "" {
			return prefix + trimmed
		}
		anchor := anchorID(pkgPath, typeName)
		if t.Config != nil && t.Config.TypeLinks != TypeLinksDisabled {
			return prefix + fmt.Sprintf("<<%s,%s>>", anchor, trimmed)
		}
		return prefix + trimmed
	}

	importPath := t.importPathForAlias(alias, file)
	if importPath == "" {
		return prefix + trimmed
	}

	if t.isInternalImport(importPath) {
		anchor := anchorID(importPath, typeName)
		if t.Config != nil && t.Config.TypeLinks != TypeLinksDisabled {
			return prefix + fmt.Sprintf("<<%s,%s>>", anchor, trimmed)
		}
		return prefix + trimmed
	}

	if t.Config != nil && t.Config.TypeLinks == TypeLinksInternalExternal {
		url := fmt.Sprintf("https://pkg.go.dev/%s#%s", importPath, typeName)
		return prefix + fmt.Sprintf("link:%s[%s]", url, trimmed)
	}

	return prefix + trimmed
}

func (t *TemplateContext) fieldSummary(field *goparser.GoField) string {
	if field == nil {
		return ""
	}
	if field.Nested != nil {
		return fmt.Sprintf("%s\tstruct", field.Nested.Name)
	}

	if t.Config == nil || t.Config.TypeLinks == TypeLinksDisabled || field.TypeInfo == nil {
		return tabifyOnce(field.Decl)
	}

	scope := t.typeParamSet(field.Struct.TypeParams)
	typeString := t.renderTypeWithScope(field.TypeInfo, scope)
	if field.Name == "" {
		if field.Tag != nil {
			return fmt.Sprintf("%s %s", typeString, field.Tag.Value)
		}
		return typeString
	}
	if field.Tag != nil {
		return fmt.Sprintf("%s\t%s %s", field.Name, typeString, field.Tag.Value)
	}
	return fmt.Sprintf("%s\t%s", field.Name, typeString)
}

func tabifyOnce(decl string) string {
	if decl == "" {
		return decl
	}
	if idx := strings.Index(decl, " "); idx != -1 {
		return decl[:idx] + "\t" + decl[idx+1:]
	}
	return decl
}

func (t *TemplateContext) fieldHeading(field *goparser.GoField) string {
	if field == nil {
		return ""
	}
	if t.Config == nil || t.Config.TypeLinks == TypeLinksDisabled || field.TypeInfo == nil {
		return field.Decl
	}
	scope := t.typeParamSet(field.Struct.TypeParams)
	typeString := t.renderTypeWithScope(field.TypeInfo, scope)
	if field.Name == "" {
		return typeString
	}
	return fmt.Sprintf("%s %s", field.Name, typeString)
}

func (t *TemplateContext) renderParameterList(params []*goparser.GoType, scope map[string]struct{}) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, param := range params {
		if param == nil {
			continue
		}
		typeString := t.renderTypeWithScope(param, scope)
		if strings.TrimSpace(param.Name) == "" {
			parts = append(parts, typeString)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", param.Name, typeString))
		}
	}
	return strings.Join(parts, ", ")
}

func (t *TemplateContext) renderResultList(results []*goparser.GoType, scope map[string]struct{}) string {
	if len(results) == 0 {
		return ""
	}
	if len(results) == 1 {
		res := results[0]
		if res == nil {
			return ""
		}
		typeString := t.renderTypeWithScope(res, scope)
		if strings.TrimSpace(res.Name) == "" {
			return typeString
		}
		return fmt.Sprintf("%s %s", res.Name, typeString)
	}
	parts := make([]string, 0, len(results))
	for _, res := range results {
		if res == nil {
			continue
		}
		typeString := t.renderTypeWithScope(res, scope)
		if strings.TrimSpace(res.Name) == "" {
			parts = append(parts, typeString)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", res.Name, typeString))
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (t *TemplateContext) functionSignatureDoc(fn *goparser.GoStructMethod) *SignatureDoc {
	segments := t.functionSignatureSegments(fn)
	if len(segments) == 0 {
		return nil
	}
	return &SignatureDoc{
		Raw:      t.functionSignature(fn),
		Kind:     SignatureKindFunction,
		Segments: segments,
	}
}

func (t *TemplateContext) methodSignatureDoc(m *goparser.GoMethod, owner []*goparser.GoType) *SignatureDoc {
	segments := t.methodSignatureSegments(m, owner)
	if len(segments) == 0 {
		return nil
	}
	return &SignatureDoc{
		Raw:      t.methodSignature(m, owner),
		Kind:     SignatureKindMethod,
		Segments: segments,
	}
}

func (t *TemplateContext) funcTypeSignatureDoc(m *goparser.GoMethod) *SignatureDoc {
	segments := t.funcTypeSignatureSegments(m)
	if len(segments) == 0 {
		return nil
	}
	return &SignatureDoc{
		Raw:      t.funcTypeSignature(m),
		Kind:     SignatureKindFuncType,
		Segments: segments,
	}
}

func (t *TemplateContext) linkedTypeSetDocs(types []*goparser.GoType) []*SignatureDoc {
	docs := []*SignatureDoc{}
	seen := map[string]struct{}{}
	scope := t.typeParamSet()
	if t.Interface != nil {
		scope = t.typeParamSet(t.Interface.TypeParams)
	}
	for _, tp := range types {
		if tp == nil {
			continue
		}
		rendered := strings.TrimSpace(t.renderTypeWithScope(tp, scope))
		if rendered == "" {
			continue
		}
		trimmed := strings.Trim(rendered, "()")
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		segments := []SignatureSegment{htmlSegment(rendered, SegmentKindText)}
		docs = append(docs, &SignatureDoc{
			Raw:      rendered,
			Kind:     SignatureKindUnknown,
			Segments: segments,
		})
	}
	return docs
}

func (t *TemplateContext) functionSignatureSegments(fn *goparser.GoStructMethod) []SignatureSegment {
	if fn == nil {
		return nil
	}
	scope := t.typeParamSet(fn.TypeParams)
	if owner := t.receiverOwnerTypeParams(fn); len(owner) > 0 {
		scope = t.typeParamSet(owner, fn.TypeParams)
	}
	segments := make([]SignatureSegment, 0, 16)
	segments = appendSegment(segments, keywordSegment("func"))
	segments = appendSegment(segments, textSegment(" "))
	if len(fn.ReceiverTypes) > 0 {
		receiver := t.htmlReceiverList(fn.ReceiverTypes, scope)
		segments = appendSegment(segments, htmlSegment("("+receiver+") ", SegmentKindReceiver))
	}
	name := strings.TrimSpace(goparser.NameWithTypeParams(fn.Name, fn.TypeParams))
	if name != "" {
		segments = appendSegment(segments, nameSegment(name))
	}
	paramsList := t.htmlParameterList(fn.Params, scope)
	params := "(" + paramsList + ")"
	if paramsList == "" {
		params = "()"
	}
	segments = appendSegment(segments, htmlSegment(params, SegmentKindParams))
	if res := t.htmlResultList(fn.Results, scope); res != "" {
		segments = appendSegment(segments, htmlSegment(" "+res, SegmentKindResult))
	}
	return segments
}

func (t *TemplateContext) methodSignatureSegments(m *goparser.GoMethod, owner []*goparser.GoType) []SignatureSegment {
	if m == nil {
		return nil
	}
	scope := t.typeParamSet(owner, m.TypeParams)
	segments := make([]SignatureSegment, 0, 12)
	name := strings.TrimSpace(goparser.NameWithTypeParams(m.Name, m.TypeParams))
	if name != "" {
		segments = appendSegment(segments, nameSegment(name))
	}
	paramsList := t.htmlParameterList(m.Params, scope)
	params := "(" + paramsList + ")"
	if paramsList == "" {
		params = "()"
	}
	segments = appendSegment(segments, htmlSegment(params, SegmentKindParams))
	if res := t.htmlResultList(m.Results, scope); res != "" {
		segments = appendSegment(segments, htmlSegment(" "+res, SegmentKindResult))
	}
	return segments
}

func (t *TemplateContext) funcTypeSignatureSegments(m *goparser.GoMethod) []SignatureSegment {
	if m == nil {
		return nil
	}
	scope := t.typeParamSet(m.TypeParams)
	segments := make([]SignatureSegment, 0, 16)
	segments = appendSegment(segments, keywordSegment("type"))
	segments = appendSegment(segments, textSegment(" "))
	name := strings.TrimSpace(goparser.NameWithTypeParams(m.Name, m.TypeParams))
	if name != "" {
		segments = appendSegment(segments, typeNameSegment(name))
	}
	segments = appendSegment(segments, textSegment(" "))
	segments = appendSegment(segments, keywordSegment("func"))
	paramsList := t.htmlParameterList(m.Params, scope)
	params := "(" + paramsList + ")"
	if paramsList == "" {
		params = "()"
	}
	segments = appendSegment(segments, htmlSegment(params, SegmentKindParams))
	if res := t.htmlResultList(m.Results, scope); res != "" {
		segments = appendSegment(segments, htmlSegment(" "+res, SegmentKindResult))
	}
	return segments
}

func appendSegment(segments []SignatureSegment, seg SignatureSegment) []SignatureSegment {
	if len(seg.Content) == 0 {
		return segments
	}
	return append(segments, seg)
}

func keywordSegment(text string) SignatureSegment {
	if text == "" {
		return SignatureSegment{}
	}
	return SignatureSegment{
		Kind:    SegmentKindKeyword,
		Content: htmltemplate.HTML(html.EscapeString(text)),
	}
}

func nameSegment(text string) SignatureSegment {
	if text == "" {
		return SignatureSegment{}
	}
	return SignatureSegment{
		Kind:    SegmentKindName,
		Content: htmltemplate.HTML(html.EscapeString(text)),
	}
}

func typeNameSegment(text string) SignatureSegment {
	if text == "" {
		return SignatureSegment{}
	}
	return SignatureSegment{
		Kind:    SegmentKindTypeName,
		Content: htmltemplate.HTML(html.EscapeString(text)),
	}
}

func textSegment(text string) SignatureSegment {
	if text == "" {
		return SignatureSegment{}
	}
	return SignatureSegment{
		Kind:    SegmentKindText,
		Content: htmltemplate.HTML(html.EscapeString(text)),
	}
}

func htmlSegment(content string, kind SignatureSegmentKind) SignatureSegment {
	if content == "" {
		return SignatureSegment{}
	}
	return SignatureSegment{
		Kind:    kind,
		Content: htmltemplate.HTML(content),
	}
}

func (t *TemplateContext) htmlReceiverList(types []*goparser.GoType, scope map[string]struct{}) string {
	parts := make([]string, 0, len(types))
	for _, gt := range types {
		if gt == nil {
			continue
		}
		segment := strings.Builder{}
		if name := strings.TrimSpace(gt.Name); name != "" {
			segment.WriteString(html.EscapeString(name))
			segment.WriteString(" ")
		}
		segment.WriteString(t.htmlType(gt, scope))
		parts = append(parts, segment.String())
	}
	return strings.Join(parts, ", ")
}

func (t *TemplateContext) htmlParameterList(params []*goparser.GoType, scope map[string]struct{}) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, param := range params {
		if param == nil {
			continue
		}
		typeString := t.htmlType(param, scope)
		if strings.TrimSpace(param.Name) == "" {
			parts = append(parts, typeString)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", html.EscapeString(param.Name), typeString))
		}
	}
	return strings.Join(parts, ", ")
}

func (t *TemplateContext) htmlResultList(results []*goparser.GoType, scope map[string]struct{}) string {
	if len(results) == 0 {
		return ""
	}
	if len(results) == 1 {
		res := results[0]
		if res == nil {
			return ""
		}
		typeString := t.htmlType(res, scope)
		if strings.TrimSpace(res.Name) == "" {
			return typeString
		}
		return fmt.Sprintf("%s %s", html.EscapeString(res.Name), typeString)
	}
	parts := make([]string, 0, len(results))
	for _, res := range results {
		if res == nil {
			continue
		}
		typeString := t.htmlType(res, scope)
		if strings.TrimSpace(res.Name) == "" {
			parts = append(parts, typeString)
		} else {
			parts = append(parts, fmt.Sprintf("%s %s", html.EscapeString(res.Name), typeString))
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func (t *TemplateContext) htmlType(gt *goparser.GoType, scope map[string]struct{}) string {
	if gt == nil {
		return ""
	}
	switch gt.Kind {
	case goparser.TypeKindPointer:
		if len(gt.Inner) > 0 {
			return "*" + t.htmlType(gt.Inner[0], scope)
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindSlice:
		if len(gt.Inner) > 0 {
			return "[]" + t.htmlType(gt.Inner[0], scope)
		}
		return "[]"
	case goparser.TypeKindArray:
		if len(gt.Inner) > 0 {
			if idx := strings.Index(gt.Type, "]"); idx != -1 {
				prefix := html.EscapeString(gt.Type[:idx+1])
				return prefix + t.htmlType(gt.Inner[0], scope)
			}
			return "[]" + t.htmlType(gt.Inner[0], scope)
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindMap:
		if len(gt.Inner) >= 2 {
			return "map[" + t.htmlType(gt.Inner[0], scope) + "]" + t.htmlType(gt.Inner[1], scope)
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindChan:
		if len(gt.Inner) > 0 {
			switch {
			case strings.HasPrefix(gt.Type, "<-chan "):
				return "<-chan " + t.htmlType(gt.Inner[0], scope)
			case strings.HasPrefix(gt.Type, "chan<- "):
				return "chan<- " + t.htmlType(gt.Inner[0], scope)
			case strings.HasPrefix(gt.Type, "chan "):
				return "chan " + t.htmlType(gt.Inner[0], scope)
			default:
				return "chan " + t.htmlType(gt.Inner[0], scope)
			}
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindEllipsis:
		if len(gt.Inner) > 0 {
			return "..." + t.htmlType(gt.Inner[0], scope)
		}
		return "..."
	case goparser.TypeKindIndex:
		if len(gt.Inner) >= 2 {
			return t.htmlType(gt.Inner[0], scope) + "[" + t.htmlType(gt.Inner[1], scope) + "]"
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindIndexList:
		if len(gt.Inner) > 0 {
			parts := make([]string, 0, len(gt.Inner)-1)
			for _, inner := range gt.Inner[1:] {
				parts = append(parts, t.htmlType(inner, scope))
			}
			return t.htmlType(gt.Inner[0], scope) + "[" + strings.Join(parts, ", ") + "]"
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindBinaryExpr:
		if len(gt.Inner) >= 2 {
			return t.htmlType(gt.Inner[0], scope) + " | " + t.htmlType(gt.Inner[1], scope)
		}
		return html.EscapeString(gt.Type)
	case goparser.TypeKindParen:
		if len(gt.Inner) > 0 {
			return "(" + t.htmlType(gt.Inner[0], scope) + ")"
		}
		return "(" + html.EscapeString(strings.Trim(gt.Type, "()")) + ")"
	case goparser.TypeKindIdent, goparser.TypeKindSelector:
		return t.htmlIdentifier(gt.Type, gt.File, scope)
	case goparser.TypeKindFunc:
		return html.EscapeString(gt.Type)
	default:
		return html.EscapeString(gt.Type)
	}
}

func (t *TemplateContext) htmlIdentifier(name string, file *goparser.GoFile, scope map[string]struct{}) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	if scope != nil {
		if _, ok := scope[trimmed]; ok {
			return html.EscapeString(trimmed)
		}
	}
	if _, ok := builtinTypes[trimmed]; ok {
		return html.EscapeString(trimmed)
	}
	alias := ""
	typeName := trimmed
	if idx := strings.Index(trimmed, "."); idx != -1 {
		alias = trimmed[:idx]
		typeName = trimmed[idx+1:]
	}
	if alias == "" {
		pkgPath := t.packagePathForFile(file)
		if pkgPath == "" {
			return html.EscapeString(trimmed)
		}
		anchor := anchorID(pkgPath, typeName)
		if t.Config != nil && t.Config.TypeLinks != TypeLinksDisabled {
			return fmt.Sprintf("<a href=\"#%s\">%s</a>", anchor, html.EscapeString(typeName))
		}
		return html.EscapeString(trimmed)
	}
	importPath := t.importPathForAlias(alias, file)
	if importPath == "" {
		return html.EscapeString(trimmed)
	}
	aliasEsc := html.EscapeString(alias)
	nameEsc := html.EscapeString(typeName)
	if t.isInternalImport(importPath) {
		anchor := anchorID(importPath, typeName)
		if t.Config != nil && t.Config.TypeLinks != TypeLinksDisabled {
			return fmt.Sprintf("%s.<a href=\"#%s\">%s</a>", aliasEsc, anchor, nameEsc)
		}
		return fmt.Sprintf("%s.%s", aliasEsc, nameEsc)
	}
	if t.Config != nil && t.Config.TypeLinks == TypeLinksInternalExternal {
		return fmt.Sprintf("%s.<a href=\"https://pkg.go.dev/%s#%s\">%s</a>", aliasEsc, importPath, typeName, nameEsc)
	}
	return fmt.Sprintf("%s.%s", aliasEsc, nameEsc)
}

func (t *TemplateContext) signatureHighlightBlocks(doc *SignatureDoc) []SignatureHighlightBlock {
	if doc == nil {
		return nil
	}
	return buildSignatureHighlightBlocks(doc)
}

func buildSignatureHighlightBlocks(doc *SignatureDoc) []SignatureHighlightBlock {
	isFuncLike := doc.Kind == SignatureKindFunction || doc.Kind == SignatureKindMethod
	blocks := make([]SignatureHighlightBlock, 0, len(doc.Segments))

	ensureBlock := func(wrapper string) *SignatureHighlightBlock {
		if len(blocks) > 0 && blocks[len(blocks)-1].WrapperClass == wrapper {
			return &blocks[len(blocks)-1]
		}
		blocks = append(blocks, SignatureHighlightBlock{WrapperClass: wrapper})
		return &blocks[len(blocks)-1]
	}

	addToken := func(wrapper, class string, content htmltemplate.HTML) {
		if len(content) == 0 {
			return
		}
		block := ensureBlock(wrapper)
		block.Tokens = append(block.Tokens, SignatureHighlightToken{
			Class:   class,
			Content: content,
		})
	}

	appendResultTokens := func(content htmltemplate.HTML) {
		if len(content) == 0 {
			return
		}
		str := string(content)
		leading := 0
		for leading < len(str) {
			switch str[leading] {
			case ' ', '\t', '\n', '\r':
				leading++
			default:
				goto whitespaceDone
			}
		}
	whitespaceDone:
		if leading > 0 {
			addToken("", "", htmltemplate.HTML(str[:leading]))
		}
		if leading < len(str) {
			addToken("", "hljs-type", htmltemplate.HTML(str[leading:]))
		}
	}

	inFunctionBlock := false

	for _, seg := range doc.Segments {
		if len(seg.Content) == 0 {
			continue
		}
		switch seg.Kind {
		case SegmentKindKeyword:
			text := strings.TrimSpace(string(seg.Content))
			wrapper := ""
			if text == "func" && (isFuncLike || doc.Kind == SignatureKindFuncType) {
				inFunctionBlock = true
				wrapper = "hljs-function"
			} else if inFunctionBlock {
				wrapper = "hljs-function"
			}
			addToken(wrapper, "hljs-keyword", seg.Content)
		case SegmentKindName:
			wrapper := ""
			if doc.Kind == SignatureKindMethod && !inFunctionBlock {
				inFunctionBlock = true
			}
			if inFunctionBlock || isFuncLike {
				wrapper = "hljs-function"
			}
			addToken(wrapper, "hljs-title", seg.Content)
		case SegmentKindTypeName:
			addToken("", "hljs-title", seg.Content)
		case SegmentKindReceiver:
			wrapper := ""
			if inFunctionBlock || doc.Kind == SignatureKindFunction {
				wrapper = "hljs-function"
			}
			addToken(wrapper, "hljs-params", seg.Content)
		case SegmentKindParams:
			wrapper := ""
			if inFunctionBlock || isFuncLike || doc.Kind == SignatureKindFuncType {
				if doc.Kind == SignatureKindFuncType && !inFunctionBlock {
					inFunctionBlock = true
				}
				wrapper = "hljs-function"
			}
			addToken(wrapper, "hljs-params", seg.Content)
		case SegmentKindResult:
			inFunctionBlock = false
			appendResultTokens(seg.Content)
		case SegmentKindText:
			wrapper := ""
			if inFunctionBlock {
				wrapper = "hljs-function"
			}
			addToken(wrapper, "", seg.Content)
		default:
			addToken("", "", seg.Content)
		}
	}

	return blocks
}

func (t *TemplateContext) typeAnchor(node interface{}) string {
	switch v := node.(type) {
	case *goparser.GoStruct:
		return t.typeAnchorFor(v.Name, v.File)
	case *goparser.GoInterface:
		return t.typeAnchorFor(v.Name, v.File)
	case *goparser.GoCustomType:
		return t.typeAnchorFor(v.Name, v.File)
	case *goparser.GoMethod:
		return t.typeAnchorFor(v.Name, v.File)
	case *goparser.GoStructMethod:
		return t.typeAnchorFor(v.Name, v.File)
	default:
		return ""
	}
}

func (t *TemplateContext) findStructByName(name string) *goparser.GoStruct {
	if t.Package == nil {
		return nil
	}
	for _, s := range t.Package.Structs {
		if s != nil && s.Name == name {
			return s
		}
	}
	return nil
}

func (t *TemplateContext) findCustomTypeByName(name string) *goparser.GoCustomType {
	if t.Package == nil {
		return nil
	}
	for _, ct := range t.Package.CustomTypes {
		if ct != nil && ct.Name == name {
			return ct
		}
	}
	return nil
}

func (t *TemplateContext) receiverOwnerTypeParams(fn *goparser.GoStructMethod) []*goparser.GoType {
	if fn == nil || len(fn.ReceiverTypes) == 0 {
		return nil
	}
	base := baseTypeIdentifier(fn.ReceiverTypes[0].Type)
	if base == "" {
		return nil
	}
	if st := t.findStructByName(base); st != nil {
		return st.TypeParams
	}
	if ct := t.findCustomTypeByName(base); ct != nil {
		return ct.TypeParams
	}
	return nil
}
func (t *TemplateContext) functionSignature(fn *goparser.GoStructMethod) string {
	if fn == nil {
		return ""
	}
	scope := t.typeParamSet(fn.TypeParams)
	if owner := t.receiverOwnerTypeParams(fn); len(owner) > 0 {
		scope = t.typeParamSet(owner, fn.TypeParams)
	}
	builder := strings.Builder{}
	builder.WriteString("func ")
	if len(fn.ReceiverTypes) > 0 {
		receivers := make([]string, 0, len(fn.ReceiverTypes))
		for _, recv := range fn.ReceiverTypes {
			if recv == nil {
				continue
			}
			typeStr := t.renderTypeWithScope(recv, scope)
			if strings.TrimSpace(recv.Name) == "" {
				receivers = append(receivers, typeStr)
			} else {
				receivers = append(receivers, fmt.Sprintf("%s %s", recv.Name, typeStr))
			}
		}
		builder.WriteString("(")
		builder.WriteString(strings.Join(receivers, ", "))
		builder.WriteString(") ")
	}
	builder.WriteString(goparser.NameWithTypeParams(fn.Name, fn.TypeParams))
	builder.WriteString("(")
	builder.WriteString(t.renderParameterList(fn.Params, scope))
	builder.WriteString(")")
	if res := t.renderResultList(fn.Results, scope); res != "" {
		builder.WriteString(" ")
		builder.WriteString(res)
	}
	return builder.String()
}

func (t *TemplateContext) methodSignature(method *goparser.GoMethod, ownerParams []*goparser.GoType) string {
	if method == nil {
		return ""
	}
	scope := t.typeParamSet(ownerParams, method.TypeParams)
	name := goparser.NameWithTypeParams(method.Name, method.TypeParams)
	params := t.renderParameterList(method.Params, scope)
	results := t.renderResultList(method.Results, scope)

	if results == "" {
		return fmt.Sprintf("%s(%s)", name, params)
	}
	return fmt.Sprintf("%s(%s) %s", name, params, results)
}

func (t *TemplateContext) funcTypeSignature(m *goparser.GoMethod) string {
	if m == nil {
		return ""
	}
	scope := t.typeParamSet(m.TypeParams)
	builder := strings.Builder{}
	builder.WriteString("type ")
	builder.WriteString(goparser.NameWithTypeParams(m.Name, m.TypeParams))
	builder.WriteString(" func(")
	builder.WriteString(t.renderParameterList(m.Params, scope))
	builder.WriteString(")")
	if res := t.renderResultList(m.Results, scope); res != "" {
		builder.WriteString(" ")
		builder.WriteString(res)
	}
	return builder.String()
}
