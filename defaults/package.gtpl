== {{if .File.FqPackage}}Package {{.File.FqPackage}}{{else}}{{.File.Decl}}{{end}}
{{if .File.BuildTags}}
[NOTE]
====
*Build Tags:* {{range $i, $tag := .File.BuildTags}}{{if $i}}, {{end}}`{{$tag}}`{{end}}
====
{{end}}

{{if (index .Docs "package-overview")}}include::{{index .Docs "package-overview"}}[leveloffset=+1]{{"\n"}}{{else}}{{ .File.Doc }}{{"\n"}}{{end}}
