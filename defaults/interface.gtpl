{{typeAnchor . .Interface}}
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
==== {{methodSignature $ . $.Interface.TypeParams}}
{{.Doc}}
{{end}}{{end}}
{{if linkedTypeSetItems . .Interface.TypeSet}}
==== Type Set
{{range linkedTypeSetItems . .Interface.TypeSet}}
* `{{.}}`
{{end}}
{{end}}
