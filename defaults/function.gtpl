=== {{nameWithTypeParams .Function.Name .Function.TypeParams}}
{{- $sig := functionSignatureDoc . .Function -}}
{{- if $sig }}
{{ renderSignature . $sig }}
{{- end }}

{{- if .Function.Doc }}
{{ .Function.Doc }}
{{- end }}

{{ if and ($sig) .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}
