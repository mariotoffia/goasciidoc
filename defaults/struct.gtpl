{{typeAnchor . .Struct}}
=== {{nameWithTypeParams .Struct.Name .Struct.TypeParams}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}{{if or .Exported $.Config.Private }}
	{{if .Nested}}{{.Nested.Name}}{{"\t"}}struct{{else}}{{tabify .Decl}}{{end}}{{end}}
{{- end}}
}
----

{{.Struct.Doc}}
{{$ctx := . -}}
{{$hasUndocumented := false -}}
{{range $field := .Struct.Fields -}}
{{- if and (or $field.Exported $ctx.Config.Private) (not $field.Nested) (not $field.Doc) }}
{{- if not $hasUndocumented }}
==== Undocumented
[cols="1,1,1",options="header"]
|===
|Field |Type |Tag
{{- $hasUndocumented = true }}
{{- end }}
|`{{ if $field.Name }}{{ $field.Name }}{{ else }}{{ $field.Decl }}{{ end }}`|`{{ if $field.Type }}{{ $field.Type }}{{ else if $field.Nested }}struct{{ else }}{{ $field.Decl }}{{ end }}`|{{ if $field.Tag }}{{ $field.Tag.Value }}{{ end }}
{{- end }}
{{- end }}
{{- if $hasUndocumented }}
|===

{{- end }}
{{range .Struct.Fields}}{{if not .Nested}}{{if or .Exported $.Config.Private }}{{if .Doc }}
==== {{fieldHeading $ .}}
{{.Doc}}
{{- end}}{{end}}{{end}}{{end}}
{{range .Struct.Fields}}{{if or .Exported $.Config.Private }}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}{{end}}
{{if hasReceivers . .Struct.Name}}{{renderReceivers . .Struct.Name}}{{end}}
