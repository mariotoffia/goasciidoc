package asciidoc

var templatePackage = `== {{ .File.Decl }}
{{ .Doc }}`

var templateImports = `== Imports
[source, go]
----
{{ declaration .File }}
----
{{range .File.Imports}}{{if .Doc }}{{ cr }}=== Import _{{ .Path }}_{{ cr }}{{ .Doc }}{{ cr }}{{end}}{{end}}`

var templateFunction = ``
var templateInterface = ``
var templateStruct = ``
var templateCustomTypeDefintion = ``
var templateCustomFuncDefintion = ``
var templateVarAssignment = ``
var templateConstAssignment = ``
