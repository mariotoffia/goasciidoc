package asciidoc

import (
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

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

func typeSetItems(types []*goparser.GoType) []string {
	items := []string{}
	seen := map[string]struct{}{}

	for _, tp := range types {
		if tp == nil {
			continue
		}

		name := strings.TrimSpace(tp.Type)
		if name == "" {
			continue
		}

		name = strings.Trim(name, "()")

		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		items = append(items, name)
	}

	return items
}
