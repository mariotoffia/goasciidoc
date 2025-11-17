{{define "module"}}
{{- if .Module}}

== Module: {{.Module.Name}}

{{- if .Module.GoVersion}}
*Go Version:* {{.Module.GoVersion}}
{{end}}
{{- if .Module.FilePath}}
*Location:* `{{.Module.Base}}`
{{end}}
{{end}}
{{- end}}