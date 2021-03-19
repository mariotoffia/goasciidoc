==== Receivers
{{range .Receiver}}{{if or .Exported $.Config.Private }}
===== {{.Name}}
[source, go]
----
{{ .Decl }}
----

{{.Doc}}
{{end}}{{end}}
