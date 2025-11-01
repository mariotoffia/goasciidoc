=== {{nameWithTypeParams .Function.Name .Function.TypeParams}}
[source, go]
----
{{ .Function.Decl }}
----

{{- $sig := functionSignature . .Function -}}
{{- if or (ne .Function.Doc "") $sig }}
{{"\n"}}
{{- if .Function.Doc }}
{{ .Function.Doc }}
{{- end }}
{{- if $sig }}
Signature:: {{ $sig }}
{{- end }}
{{- end }}
{{ if .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}
