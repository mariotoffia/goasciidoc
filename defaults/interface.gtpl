=== {{nameWithTypeParams .Interface.Name .Interface.TypeParams}}
[source, go]
----
{{.Interface.Decl}} {
{{- range .Interface.TypeSet}}
	{{.Type}}
{{- end}}
{{- range .Interface.Methods}}{{if or .Exported $.Config.Private }}
	{{tabifylast .Decl}}{{end}}
{{- end}}
}
----
	
{{.Interface.Doc}}
{{range .Interface.Methods}}{{if or .Exported $.Config.Private }}
==== {{.Decl}}
{{.Doc}}
{{end}}{{end}}
{{if .Interface.TypeSet}}
==== Type Set
{{range .Interface.TypeSet}}
* `{{.Type}}`
{{end}}
{{end}}
