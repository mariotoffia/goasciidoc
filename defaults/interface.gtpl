=== {{nameWithTypeParams .Interface.Name .Interface.TypeParams}}
[source, go]
----
{{.Interface.Decl}} {
{{- range .Interface.TypeSetDecl}}
	{{.}}
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
{{if typeSetItems .Interface.TypeSet}}
==== Type Set
{{range typeSetItems .Interface.TypeSet}}
* `{{.}}`
{{end}}
{{end}}
