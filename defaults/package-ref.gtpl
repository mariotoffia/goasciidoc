{{define "package-ref"}}
{{- if .Package}}
{{- if .PackageAnchor}}

[[{{.PackageAnchor}}]]
{{- end}}
== Package: {{.Package.Name}}

{{- if .Config.PackageModeInclude}}
{{- if .PackageFile}}

include::{{.PackageFile}}[]
{{- end}}
{{- else}}
{{- if .Package.Doc}}

{{.Package.Doc}}
{{end}}
{{- if .PackageFile}}

link:{{.PackageFile}}[View full documentation]
{{end}}
{{- end}}
{{end}}
{{- end}}
