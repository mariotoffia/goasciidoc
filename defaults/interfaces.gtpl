== Interfaces

{{range .File.Interfaces}}{{if or .Exported $.Config.Private }}
{{- render $ .}}{{end}}
{{end}}
