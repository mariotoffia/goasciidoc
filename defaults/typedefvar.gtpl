{{typeAnchor . .TypeDefVar}}
=== {{nameWithTypeParams .TypeDefVar.Name .TypeDefVar.TypeParams}}
[source, go]
----
{{.TypeDefVar.Decl}}
----

{{.TypeDefVar.Doc}}

{{if hasReceivers . .TypeDefVar.Name}}{{renderReceivers . .TypeDefVar.Name}}{{end}}
