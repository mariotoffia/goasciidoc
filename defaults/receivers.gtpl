==== Receivers
{{range .Receiver}}{{if or .Exported $.Config.Private }}
===== {{nameWithTypeParams .Name .TypeParams}}
{{- $sig := functionSignatureDoc $ . -}}
{{- if $sig }}
{{ renderSignature $ $sig }}
{{- end }}

{{- if .Doc }}
{{.Doc}}
{{- end }}

{{end}}{{end}}
