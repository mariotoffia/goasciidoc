[#{{.Function.File.FqPackage}}/{{.Function.Name}}]
=== {{ .Function.Name }}
[source, go]
----
{{ .Function.Decl }}
----

{{ .Function.Doc }}
{{ if .Config.IncludeMethodCode }}{{"\n"}}[source, go]{{"\n"}}----{{"\n"}}{{ .Function.FullDecl }}{{"\n"}}----{{end}}
