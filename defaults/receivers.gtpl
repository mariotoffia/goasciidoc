==== Receivers
{{range .Receiver}}{{if or .Exported $.Config.Private }}
===== {{nameWithTypeParams .Name .TypeParams}}
[source, go]
----
{{ .Decl }}
----

{{- if .Doc }}
{{.Doc}}
{{- end }}
{{- $sigHTML := functionSignatureHTML $ . -}}
{{- if $sigHTML }}
+++
<div class="listingblock signature">
<div class="content">
<pre class="highlight"><code class="language-go">{{$sigHTML}}</code></pre>
</div>
</div>
+++
{{- end }}
{{end}}{{end}}
