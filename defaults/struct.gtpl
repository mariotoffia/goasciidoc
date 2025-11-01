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
{{range .Struct.Fields}}{{if not .Nested}}{{if or .Exported $.Config.Private }}
==== {{fieldHeading $ .}}
{{.Doc}}
{{- end}}
{{end}}{{end}}
{{range .Struct.Fields}}{{if or .Exported $.Config.Private }}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}{{end}}
{{if hasReceivers . .Struct.Name}}{{renderReceivers . .Struct.Name}}{{end}}
