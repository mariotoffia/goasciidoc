{{- if or .PackageRefs.Internal .PackageRefs.External}}

== Package References

{{- if .PackageRefs.Internal}}

=== Internal Packages

This package references the following packages within this project:

{{range .PackageRefs.Internal}}
* {{if .File}}link:{{.File}}[{{.Name}}]{{else if .Anchor}}<<{{.Anchor}},{{.Name}}>>{{else}}{{.Name}}{{end}}
{{end}}
{{- end}}

{{- if .PackageRefs.External}}

=== External Packages

This package imports the following external packages:

{{range .PackageRefs.External}}
* {{if .File}}link:{{.File}}[{{.Name}}]{{else}}`{{.Name}}`{{end}}{{if .Doc}} - {{.Doc}}{{end}}
{{- end}}
{{- end}}
{{- end}}
