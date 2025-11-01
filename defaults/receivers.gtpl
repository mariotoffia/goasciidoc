==== Receivers
{{range .Receiver}}{{if or .Exported $.Config.Private }}
===== {{nameWithTypeParams .Name .TypeParams}}
[source, go]
----
{{ .Decl }}
----

{{- $sig := functionSignature $ . -}}
{{- if or (ne .Doc "") $sig }}
{{"\n"}}
{{- if .Doc }}
{{.Doc}}
{{- end }}
{{- if $sig }}
Signature:: {{ $sig }}
{{- end }}
{{- end }}
{{end}}{{end}}
