{{- if or .PackageRefs.Internal .PackageRefs.External}}

== Package References

{{- if .PackageRefs.Internal}}

=== Internal Packages

This package references the following packages within this project:

{{- range .PackageRefs.Internal}}
* <<{{.Anchor}},{{.Name}}>>{{if .File}} - link:{{.File}}[Documentation]{{end}}
{{- end}}
{{- end}}

{{- if .PackageRefs.External}}

=== External Packages

This package imports the following external packages:

{{- range .PackageRefs.External}}
* `{{.Name}}`{{if .Doc}} - {{.Doc}}{{end}}
{{- end}}
{{- end}}
{{- end}}
