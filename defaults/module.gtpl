{{define "module"}}
{{- if .Module}}
{{- if .ModuleAnchor}}

[[{{.ModuleAnchor}}]]
{{- end}}
== Module: {{.Module.Name}}

{{- if .Config.ModuleModeInclude}}
{{- if .ModuleFile}}

include::{{.ModuleFile}}[]
{{- end}}
{{- else}}
{{- if .Module.GoVersion}}

*Go Version:* {{.Module.GoVersion}}
{{end}}
{{- if .Module.FilePath}}
*Location:* `{{.Module.Base}}`
{{end}}
{{- end}}
{{end}}
{{- end}}