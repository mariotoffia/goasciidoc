{{- if .Index}}{{/* Standalone package mode - full document header */}}
= {{if .File.FqPackage}}Package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}
{{- if .Index.AuthorName}}{{"\n"}}:author_name: {{.Index.AuthorName}}{{"\n"}}:author: {author_name}{{end}}
{{- if .Index.AuthorEmail}}{{"\n"}}:author_email: {{.Index.AuthorEmail}}{{"\n"}}:email: {author_email}{{end}}
:source-highlighter: {{ .Index.Highlighter }}
{{- if .Index.TocTitle}}{{"\n"}}:toc:{{"\n"}}:toc-title: {{ .Index.TocTitle }}{{"\n"}}:toclevels: {{ .Index.TocLevels }}{{end}}
:icons: font
{{- if .Index.ImageDir}}{{"\n"}}:imagesdir: {{.Index.ImageDir}}{{end}}
{{- if .Index.HomePage}}{{"\n"}}:homepage: {{.Index.HomePage}}{{end}}
:kroki-default-format: svg
:doctype: {{.Index.DocType}}
{{- else -}}{{/* Regular inline mode - section heading only */}}
== {{if .File.FqPackage}}Package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}
{{- end}}
{{if .File.BuildTags}}
[NOTE]
====
*Build Tags:* {{range $i, $tag := .File.BuildTags}}{{if $i}}, {{end}}`{{$tag}}`{{end}}
====
{{end}}

{{if (index .Docs "package-overview")}}include::{{index .Docs "package-overview"}}[leveloffset=+1]{{"\n"}}{{else}}{{ processReferences . .File.Doc }}{{"\n"}}{{end}}
