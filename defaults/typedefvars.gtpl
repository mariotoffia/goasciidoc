== Variable Typedefinitions

{{range .File.CustomTypes}}{{if or .Exported $.Config.Private }}
{{- render $ .}}{{end}}
{{end}}
