== Structs

{{range .File.Structs}}{{if or .Exported $.Config.Private }}
{{- render $ .}}{{end}}
{{end}}
