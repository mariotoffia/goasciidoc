{{typeAnchor . .TypeDefFunc}}
=== {{nameWithTypeParams .TypeDefFunc.Name .TypeDefFunc.TypeParams}}
[source, go]
----
{{.TypeDefFunc.Decl}}
----
{{.TypeDefFunc.Doc}}
{{- $sigHTML := funcTypeSignatureHTML . .TypeDefFunc -}}
{{- if $sigHTML }}
+++
<div class="listingblock signature">
<div class="content">
<pre class="highlight"><code class="language-go">{{$sigHTML}}</code></pre>
</div>
</div>
+++
{{- end }}
