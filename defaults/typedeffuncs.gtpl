== Function Definitions

{{range .File.CustomFuncs}}{{if or .Exported $.Config.Private }}
{{render $ .}}{{end}}
{{end}}
