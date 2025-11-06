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
{{- $ifaceDoc := trimnl .Interface.Doc -}}
{{if $ifaceDoc}}
{{printf "\n%s\n\n" $ifaceDoc}}
{{else}}
{{printf "\n"}}
{{end}}

{{- $ctx := . -}}
{{- $hasUndocumented := false -}}
{{- range $method := .Interface.Methods}}
{{- if and (or $method.Exported $ctx.Config.Private) (not $method.Doc) }}
{{- if not $hasUndocumented}}
{{printf "==== Undocumented\n\n"}}
[cols="1,1",options="header"]
|===
|Name |Signature
{{- $hasUndocumented = true }}
{{- end}}
|`{{ $method.Name }}`|`{{ $method.Decl }}`
{{- end}}
{{- end}}
{{- if $hasUndocumented }}
|===

{{- printf "\n" -}}
{{- end}}
{{- range .Interface.Methods}}{{- if or .Exported $.Config.Private }}
{{- $doc := trimnl .Doc -}}
{{- if $doc }}
{{- $sig := methodSignatureDoc $ . $.Interface.TypeParams -}}
{{- $style := $.Config.SignatureStyle -}}
==== {{if $sig}}{{ $sig.Raw }}{{else}}{{ .Decl }}{{end}}
{{printf "\n"}}
{{printf "%s\n\n" $doc}}
{{if $sig }}
{{if eq $style "goasciidoc" }}
{{- $blocks := signatureHighlightBlocks $ $sig -}}
{{if gt (len $blocks) 0 }}
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

{{printf "\n"}}
{{end }}
{{else }}
[source, go]
----
{{signaturePlain $ $sig}}
----

{{printf "\n"}}
{{end }}
{{end }}
{{end }}
{{- end}}{{- end}}
{{- $typeDocs := linkedTypeSetDocs . .Interface.TypeSet -}}
{{if $typeDocs}}
{{printf "==== Type Set\n\n"}}
{{range $typeDocs}}
* `{{- range .Segments -}}{{ .Content }}{{- end -}}`
{{end}}
{{end}}
