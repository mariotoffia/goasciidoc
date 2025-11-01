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
{{- if .Doc }}
{{.Doc}}
{{- end }}
{{- $sig := methodSignatureDoc $ . $.Interface.TypeParams -}}
{{- if $sig }}
{{ renderSignature $ $sig }}
{{- end }}
{{end}}{{end}}
{{if linkedTypeSetItems . .Interface.TypeSet}}
==== Type Set
{{range linkedTypeSetItems . .Interface.TypeSet}}
* `{{.}}`
{{end}}
{{end}}
