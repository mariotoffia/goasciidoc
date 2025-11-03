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
{{- $sig := methodSignatureDoc $ . $.Interface.TypeParams -}}
{{- $style := $.Config.SignatureStyle -}}
==== {{ if and $sig (eq $style "highlight") -}}
{{- $blocks := signatureHighlightBlocks $ $sig -}}
{{- range $block := $blocks -}}
{{- if $block.WrapperClass }}<span class="{{ $block.WrapperClass }}">{{- end -}}
{{- range $token := $block.Tokens -}}
{{- if $token.Class }}<span class="{{ $token.Class }}">{{- end -}}{{ $token.Content }}{{- if $token.Class }}</span>{{- end -}}
{{- end -}}
{{- if $block.WrapperClass }}</span>{{- end -}}
{{- end -}}
{{- else if $sig -}}
{{- range $seg := $sig.Segments -}}{{ $seg.Content }}{{- end -}}
{{- else -}}
{{ .Decl }}
{{- end -}}
{{- if .Doc }}
{{.Doc}}
{{- end }}
{{- if $sig }}
{{- if eq $style "highlight" }}
{{- $blocks := signatureHighlightBlocks $ $sig -}}
{{- if gt (len $blocks) 0 }}
++++
<div class="listingblock signature">
<div class="content">
<pre class="highlightjs highlight"><code class="language-go hljs">{{- range $block := $blocks -}}
{{- if $block.WrapperClass }}<span class="{{ $block.WrapperClass }}">{{- end -}}
{{- range $token := $block.Tokens -}}
{{- if $token.Class }}<span class="{{ $token.Class }}">{{- end -}}{{ $token.Content }}{{- if $token.Class }}</span>{{- end -}}
{{- end -}}
{{- if $block.WrapperClass }}</span>{{- end -}}
{{- end }}</code></pre>
</div>
</div>
++++

{{- end }}
{{- else if $sig.Raw }}
[source, go]
----
{{$sig.Raw}}
----

{{- end }}
{{ end }}
{{end}}{{end}}
{{- $typeDocs := linkedTypeSetDocs . .Interface.TypeSet -}}
{{if $typeDocs}}
==== Type Set
{{range $typeDocs}}
* `{{- range .Segments -}}{{ .Content }}{{- end -}}`
{{end}}
{{end}}
