=== {{nameWithTypeParams .Function.Name .Function.TypeParams}}
[source, go]
----
{{ .Function.Decl }}
----

{{- $sigHTML := functionSignatureHTML . .Function -}}
{{- if .Function.Doc }}
{{ .Function.Doc }}
{{- end }}
{{- if $sigHTML }}
+++
<div class="listingblock signature">
<div class="content">
<pre class="highlight"><code class="language-go">{{$sigHTML}}</code></pre>
</div>
</div>
+++
{{- end }}

{{ if .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}
