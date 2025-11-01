{{typeAnchor . .TypeDefFunc}}
=== {{nameWithTypeParams .TypeDefFunc.Name .TypeDefFunc.TypeParams}}
{{- $sig := funcTypeSignatureDoc . .TypeDefFunc -}}
{{- if $sig }}
{{ renderSignature . $sig }}
{{- end }}
{{.TypeDefFunc.Doc}}
