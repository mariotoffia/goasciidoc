==== Receivers
{{range .Receiver}}{{if or .Exported $.Config.Private }}
===== {{nameWithTypeParams .Name .TypeParams}}
[source, go]
----
{{ .Decl }}
----

{{.Doc}}
{{end}}{{end}}
