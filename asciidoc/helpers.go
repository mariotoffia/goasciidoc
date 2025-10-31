package asciidoc

import "github.com/mariotoffia/goasciidoc/goparser"

func typeParamsSuffix(params []*goparser.GoType) string {
	return goparser.FormatTypeParams(params)
}

func nameWithTypeParams(name string, params []*goparser.GoType) string {
	return goparser.NameWithTypeParams(name, params)
}

func indent(line string) string {
	if line == "" {
		return ""
	}
	return "\t" + line
}
