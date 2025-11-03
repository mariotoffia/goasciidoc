=== {{nameWithTypeParams .Function.Name .Function.TypeParams}}
{{- $sig := functionSignatureDoc . .Function -}}
{{- if $sig }}
{{- $style := .Config.SignatureStyle }}
{{- if eq $style "highlight" }}
{{- $blocks := signatureHighlightBlocks . $sig -}}
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
{{- end }}

{{- if .Function.Doc }}
{{ .Function.Doc }}
{{- end }}

{{ if and ($sig) .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}
