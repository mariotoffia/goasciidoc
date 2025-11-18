== Imports
[source, go]
----
{{ render . }}
----
{{range .File.Imports}}{{if .Doc }}{{"\n"}}==== Import _{{ .Path }}_{{"\n"}}{{ .Doc }}{{"\n"}}{{end}}{{end}}
