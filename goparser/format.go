package goparser

import (
	"strings"
)

// FormatTypeParams returns the textual representation of a list of type parameters.
// For example, `[T any, U ~string]` or an empty string when there are no type parameters.
func FormatTypeParams(params []*GoType) string {

	if len(params) == 0 {
		return ""
	}

	parts := make([]string, 0, len(params))
	for _, param := range params {

		if param == nil {
			continue
		}

		name := strings.TrimSpace(param.Name)
		constraint := strings.TrimSpace(param.Type)

		switch {
		case name == "":
			continue
		case constraint == "":
			parts = append(parts, name)
		default:
			parts = append(parts, name+" "+constraint)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// NameWithTypeParams returns the identifier including formatted type parameters.
func NameWithTypeParams(name string, params []*GoType) string {
	return strings.TrimSpace(name) + FormatTypeParams(params)
}
