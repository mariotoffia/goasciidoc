package asciidoc

var templatePackage = `==  {{ declaration . }}
{{ .Doc }}`

var templateImports = `== Imports
[source, go]
----
{{ declaration .File }}
----

{{range .File.Imports}}{{if .Doc }}=== Import _{{ .Path }}_{{ cr }}{{ .Doc }}{{ cr }}{{ cr }}{{end}}{{end}}`

var templateFunction = ``
var templateInterface = ``
var templateStruct = ``
var templateCustomTypeDefintion = ``
var templateCustomFuncDefintion = ``
var templateVarAssignment = ``
var templateConstAssignment = ``
