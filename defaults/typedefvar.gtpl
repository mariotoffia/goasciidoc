=== {{.TypeDefVar.Name}}
[source, go]
----
{{.TypeDefVar.Decl}}
----

{{.TypeDefVar.Doc}}

{{if hasReceivers . .TypeDefVar.Name}}{{renderReceivers . .TypeDefVar.Name}}{{end}}
