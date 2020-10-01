package asciidoc

var templatePackage = `== {{if .File.FqPackage}}package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}
{{ .File.Doc }}`

var templateImports = `=== Imports
[source, go]
----
{{ render . }}
----
{{range .File.Imports}}{{if .Doc }}{{ cr }}==== Import _{{ .Path }}_{{ cr }}{{ .Doc }}{{ cr }}{{end}}{{end}}`

var templateFunctions = `== Functions
{{range .File.StructMethods}}
{{- render $ .}}
{{end}}`

var templateFunction = `=== {{ .Function.Name }}
[source, go]
----
{{ .Function.Decl }}
----

{{ .Function.Doc }}
{{ if .Config.IncludeMethodCode }}{{cr}}[source, go]{{cr}}----{{cr}}{{ .Function.FullDecl }}{{cr}}----{{end}}`

var templateInterface = `=== {{ .Interface.Name }}
[source, go]
----
{{.Interface.Decl}} {
{{- range .Interface.Methods}}
	{{.Decl}}
{{- end}}
}
----
		
{{ .Interface.Doc }}
{{range .Interface.Methods}}
==== {{.Decl}}
{{.Doc}}
{{end}}`

var templateInterfaces = `== Interfaces
{{range .File.Interfaces}}
{{- render $ .}}
{{end}}`

var templateStruct = `=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
	{{.Decl}}
{{- end}}
}
----
		
{{ .Struct.Doc }}
{{range .Struct.Fields}}
==== {{.Decl}}
{{.Doc}}
{{end}}`

var templateStructs = `== Structs
{{range .File.Structs}}
{{- render $ .}}
{{end}}`

var templateCustomTypeDefintion = `=== {{.TypeDefVar.Name}}
[source, go]
----
{{.TypeDefVar.Decl}}
----
{{.TypeDefVar.Doc}}`

var templateCustomTypeDefintions = `== Variable Typedefinitions
{{range .File.CustomTypes}}
{{- render $ .}}
{{end}}`

var templateVarAssignment = `=== {{.VarAssignment.Name}}
[source, go]
----
{{.VarAssignment.FullDecl}}
----
{{.VarAssignment.Doc}}`

var templateVarAssignments = `== Variables
{{range .File.VarAssigments}}
{{render $ .}}
{{end}}`

var templateConstAssignment = `=== {{.ConstAssignment.Name}}
[source, go]
----
{{.ConstAssignment.Decl}}
----
{{.ConstAssignment.Doc}}`

var templateConstAssignments = `=== Constants
[source, go]
----
const (
	{{- range .File.ConstAssignments}}
	{{.Decl}}
	{{- end}}
)
----
{{range .File.ConstAssignments}}
{{render $ .}}
{{end}}`

var templateCustomFuncDefintion = `=== {{.TypeDefFunc.Name}}
[source, go]
----
{{.TypeDefFunc.Decl}}
----
{{.TypeDefFunc.Doc}}`

var templateCustomFuncDefintions = ``
