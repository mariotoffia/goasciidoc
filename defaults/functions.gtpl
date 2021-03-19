== Functions

{{range .File.StructMethods}}
{{- if notreceiver $ .}}{{if or .Exported $.Config.Private }}{{render $ .}}{{end}}{{end}}
{{end}}
