= {{ .Index.Title }}
{{- if .Index.AuthorName}}{{"\n"}}:author_name: {{.Index.AuthorName}}{{"\n"}}:author: {author_name}{{end}}
{{- if .Index.AuthorEmail}}{{"\n"}}:author_email: {{.Index.AuthorEmail}}{{"\n"}}:email: {author_email}{{end}}
:source-highlighter: {{ .Index.Highlighter }}
{{- if .Index.TocTitle}}{{"\n"}}:toc:{{"\n"}}:toc-title: {{ .Index.TocTitle }}{{"\n"}}:toclevels: {{ .Index.TocLevels }}{{end}}
:icons: font
{{- if .Index.ImageDir}}{{"\n"}}:imagesdir: {{.Index.ImageDir}}{{end}}
{{- if .Index.HomePage}}{{"\n"}}:homepage: {{.Index.HomePage}}{{end}}
:kroki-default-format: svg
:doctype: {{.Index.DocType}}

{{- if .Workspace}}

== Overview

This documentation covers the following modules:

{{- range $i, $module := .Workspace.Modules}}
{{ add $i 1 }}. <<module-{{ add $i 1 }},{{$module.Name}}>>
{{- end}}
{{- end}}

