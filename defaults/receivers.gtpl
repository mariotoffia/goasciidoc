==== Receivers
{{range .Receiver}}{{if or .Exported $.Config.Private }}
===== {{nameWithTypeParams .Name .TypeParams}}
{{- $sig := functionSignatureDoc $ . -}}
{{- if $sig }}
{{- $style := $.Config.SignatureStyle }}
{{- if eq $style "goasciidoc" }}
{{- $blocks := signatureHighlightBlocks $ $sig -}}
{{- if gt (len $blocks) 0 }}
+++
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
+++

{{- end }}
{{- else }}
[source, go]
----
{{signaturePlain $ $sig}}
----

{{- end }}
{{- end }}

{{- if .Doc }}
{{processReferences $ .Doc}}
{{- end }}

{{end}}{{end}}
