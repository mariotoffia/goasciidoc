{{typeAnchor . .Struct}}
=== {{nameWithTypeParams .Struct.Name .Struct.TypeParams}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}{{if or .Exported $.Config.Private }}
	{{if .AnonymousStruct}}{{.AnonymousStruct.Name}}{{"\t"}}struct{{else}}{{tabify .Decl}}{{end}}{{end}}
{{- end}}
}
----
{{- $structDoc := trimnl .Struct.Doc -}}
{{if $structDoc}}
{{printf "\n%s\n\n" $structDoc}}
{{else}}
{{printf "\n"}}
{{end}}
{{- $shouldRenderJSON := false -}}
{{- $shouldRenderYAML := false -}}
{{- if .Config.RenderOptions -}}
  {{- if index .Config.RenderOptions "struct-json" -}}
    {{- $shouldRenderJSON = true -}}
  {{- end -}}
  {{- if index .Config.RenderOptions "struct-yaml" -}}
    {{- $shouldRenderYAML = true -}}
  {{- end -}}
{{- else -}}
  {{- $shouldRenderJSON = true -}}
  {{- $shouldRenderYAML = true -}}
{{- end -}}
{{- if and $shouldRenderJSON (hasJSONTag .Struct)}}
==== JSON Example
[source, json]
----
{{toJSON .Struct}}
----

{{end}}
{{- if and $shouldRenderYAML (hasYAMLTag .Struct)}}
==== YAML Example
[source, yaml]
----
{{toYAML .Struct}}
----

{{end}}

{{- $ctx := . -}}
{{- $hasUndocumented := false -}}
{{- range $field := .Struct.Fields}}
{{- if and (or $field.Exported $ctx.Config.Private) (not $field.AnonymousStruct) (not $field.Doc) }}
{{- if not $hasUndocumented}}
{{printf "==== Undocumented\n\n"}}
[cols="1,1,1",options="header"]
|===
|Field |Type |Tag
{{- $hasUndocumented = true }}
{{- end}}
|`{{ if $field.Name }}{{ $field.Name }}{{ else }}{{ $field.Decl }}{{ end }}`|`{{ if $field.Type }}{{ $field.Type }}{{ else if $field.AnonymousStruct }}struct{{ else }}{{ $field.Decl }}{{ end }}`|{{ if $field.Tag }}{{ $field.Tag.Value }}{{ end }}
{{- end}}
{{- end}}
{{- if $hasUndocumented }}
|===

{{- end}}
{{- range .Struct.Fields}}
{{- if not .AnonymousStruct}}
{{- if or .Exported $.Config.Private }}
{{- $doc := trimnl .Doc -}}
{{- if $doc }}
{{printf "==== %s\n\n" (fieldHeading $ .)}}
{{printf "%s\n\n" $doc}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{range .Struct.Fields}}{{if or .Exported $.Config.Private }}{{if .AnonymousStruct}}{{render $ .AnonymousStruct}}{{end}}{{end}}{{end}}
{{if hasReceivers . .Struct.Name}}{{renderReceivers . .Struct.Name}}{{end}}
