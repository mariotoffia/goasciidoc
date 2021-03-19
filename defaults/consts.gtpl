=== Constants
[source, go]
----
const (
	{{- range .File.ConstAssignments}}{{if or .Exported $.Config.Private }}
	{{tabify .Decl}}{{end}}
	{{- end}}
)
----
{{range .File.ConstAssignments}}{{if or .Exported $.Config.Private }}
{{render $ .}}{{end}}
{{end}}
