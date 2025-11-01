{{ $doc := .Doc }}
{{ if eq .Style "highlight" }}
{{ if gt (len $doc.Segments) 0 }}
++++
<div class="listingblock signature">
<div class="content">
<pre class="highlightjs highlight"><code class="language-go hljs">{{ range $doc.Segments }}{{ if .Class }}<span class="{{.Class}}">{{.Content}}</span>{{ else }}{{.Content}}{{ end }}{{ end }}</code></pre>
</div>
</div>
++++

{{ end }}
{{ else if $doc.Raw }}
[source, go]
----
{{$doc.Raw}}
----
{{ end }}
