{{define "package-ref"}}
{{- if .Package}}
{{- if .Config.PackageModeInclude}}
{{- if .PackageFile}}
include::{{.PackageFile}}[leveloffset=+1]
{{- end}}
{{- else}}
{{- if .PackageAnchor}}

[[{{.PackageAnchor}}]]
{{- end}}
== Package: {{if .Package.FqPackage}}{{.Package.FqPackage}}{{else}}{{.Package.Package}}{{end}}

{{- if .Package.Doc}}

{{.Package.Doc}}
{{end}}
{{- if .PackageFile}}

link:{{.PackageFile}}[View full documentation]
{{end}}
{{- end}}
{{end}}
{{- end}}
