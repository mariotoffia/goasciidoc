package asciidoc

var templateIndex = `= {{ .Index.Title }}
{{- if .Index.AuthorName}}{{"\n"}}:author_name: {{.Index.AuthorName}}{{"\n"}}:author: {author_name}{{end}}
{{- if .Index.AuthorEmail}}{{"\n"}}:author_email: {{.Index.AuthorEmail}}{{"\n"}}:email: {author_email}{{end}}
:source-highlighter: {{ .Index.Highlighter }}
{{- if .Index.TocTitle}}{{"\n"}}:toc:{{"\n"}}:toc-title: {{ .Index.TocTitle }}{{"\n"}}:toclevels: {{ .Index.TocLevels }}{{end}}
:icons: font
{{- if .Index.ImageDir}}{{"\n"}}:imagesdir: {{.Index.ImageDir}}{{end}}
{{- if .Index.HomePage}}{{"\n"}}:homepage: {{.Index.HomePage}}{{end}}
:kroki-default-format: svg
:doctype: {{.Index.DocType}}

`

var templatePackage = `== {{if .File.FqPackage}}Package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}

{{if (index .Docs "package-overview")}}include::{{index .Docs "package-overview"}}[leveloffset=+1]{{"\n"}}{{else}}{{ .File.Doc }}{{end}}
`

var templateImports = `=== Imports
[source, go]
----
{{ render . }}
----
{{range .File.Imports}}{{if .Doc }}{{"\n"}}==== Import _{{ .Path }}_{{"\n"}}{{ .Doc }}{{"\n"}}{{end}}{{end}}
`

var templateFunctions = `== Functions

{{range .File.StructMethods}}
{{- if notreceiver $ .}}{{render $ .}}{{end}}
{{end}}
`

var templateFunction = `=== {{ .Function.Name }}
[source, go]
----
{{ .Function.Decl }}
----

{{ .Function.Doc }}
{{ if .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}
`

var templateInterface = `=== {{ .Interface.Name }}
[source, go]
----
{{.Interface.Decl}} {
{{- range .Interface.Methods}}
	{{tabifylast .Decl}}
{{- end}}
}
----
		
{{.Interface.Doc}}
{{range .Interface.Methods}}
==== {{.Decl}}
{{.Doc}}
{{end}}
`

var templateInterfaces = `== Interfaces

{{range .File.Interfaces}}
{{- render $ .}}
{{end}}
`

var templateStruct = `=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
	{{if .Nested}}{{.Nested.Name}}{{"\t"}}struct{{else}}{{tabify .Decl}}{{end}}
{{- end}}
}
----

{{.Struct.Doc}}
{{range .Struct.Fields}}{{if not .Nested}}
==== {{.Decl}}
{{.Doc}}
{{- end}}
{{end}}
{{range .Struct.Fields}}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}
{{if hasReceivers . .Struct.Name}}{{renderReceivers . .Struct.Name}}{{end}}
`

var templateStructs = `== Structs

{{range .File.Structs}}
{{- render $ .}}
{{end}}
`

var templateReceivers = `==== Receivers
{{range .Receiver}}
===== {{.Name}}
[source, go]
----
{{ .Decl }}
----

{{.Doc}}
{{end}}
`

var templateCustomTypeDefintion = `=== {{.TypeDefVar.Name}}
[source, go]
----
{{.TypeDefVar.Decl}}
----

{{.TypeDefVar.Doc}}

{{if hasReceivers . .TypeDefVar.Name}}{{renderReceivers . .TypeDefVar.Name}}{{end}}
`

var templateCustomTypeDefintions = `== Variable Typedefinitions

{{range .File.CustomTypes}}
{{- render $ .}}
{{end}}
`

var templateVarAssignment = `=== {{.VarAssignment.Name}}
[source, go]
----
{{.VarAssignment.FullDecl}}
----
{{.VarAssignment.Doc}}
`

var templateVarAssignments = `== Variables
{{range .File.VarAssignments}}
{{render $ .}}
{{end}}
`

var templateConstAssignment = `=== {{.ConstAssignment.Name}}
[source, go]
----
{{.ConstAssignment.Decl}}
----
{{.ConstAssignment.Doc}}
`

var templateConstAssignments = `=== Constants
[source, go]
----
const (
	{{- range .File.ConstAssignments}}
	{{tabify .Decl}}
	{{- end}}
)
----
{{range .File.ConstAssignments}}
{{render $ .}}
{{end}}
`

var templateCustomFuncDefintion = `=== {{.TypeDefFunc.Name}}
[source, go]
----
{{.TypeDefFunc.Decl}}
----
{{.TypeDefFunc.Doc}}
`

var templateCustomFuncDefintions = `== Function Definitions

{{range .File.CustomFuncs}}
{{render $ .}}
{{end}}
`
