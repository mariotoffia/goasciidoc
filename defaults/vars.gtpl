== Variables
{{range .File.VarAssignments}}{{if or .Exported $.Config.Private }}
{{render $ .}}{{end}}
{{end}}
