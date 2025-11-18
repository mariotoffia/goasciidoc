{{define "package-ref"}}
{{- if .Package}}
{{- if .PackageAnchor}}

[[{{.PackageAnchor}}]]
{{- end}}
== Package: {{if .Package.FqPackage}}{{.Package.FqPackage}}{{else}}{{.Package.Package}}{{end}}

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
