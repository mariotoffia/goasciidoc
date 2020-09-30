package asciidoc

var templatePackage = `== {{ .File.Decl }}
{{ .Doc }}`

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

var templateInterface = ``
var templateStruct = ``
var templateCustomTypeDefintion = ``
var templateCustomFuncDefintion = ``
var templateVarAssignment = ``
var templateConstAssignment = ``
