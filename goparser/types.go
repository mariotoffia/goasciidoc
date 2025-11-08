package goparser

import "strings"

// GoAssignment represents a single var assignment e.g. var pelle = 10
type GoAssignment struct {
	File *GoFile
	Name string
	Doc  string
	// Decl will be the same if multi var assignment on same row e.g. var pelle, lisa = 10, 19
	// then both pelle and list will have 'var pelle, lisa = 10, 19' as Decl
	Decl     string
	FullDecl string
	Exported bool
}

// GoCustomType is a custom type definition
type GoCustomType struct {
	File       *GoFile
	Name       string
	Doc        string
	Type       string
	Decl       string
	Exported   bool
	TypeParams []*GoType
}

// GoInterface specifies a interface definition
type GoInterface struct {
	File        *GoFile
	Doc         string
	Decl        string
	FullDecl    string
	Name        string
	Exported    bool
	Methods     []*GoMethod
	TypeParams  []*GoType
	TypeSet     []*GoType
	TypeSetDecl []string
}

// GoType represents a go type such as a array, map, custom type etc.
type GoType struct {
	File       *GoFile
	Name       string
	Type       string
	Underlying string
	Exported   bool
	Inner      []*GoType
	Kind       TypeKind
}

// GoStruct represents a struct
type GoStruct struct {
	File       *GoFile
	Doc        string
	Decl       string
	FullDecl   string
	Name       string
	Exported   bool
	Fields     []*GoField
	TypeParams []*GoType
}

// HasJSONTag returns true if any field in the struct has a json tag
func (s *GoStruct) HasJSONTag() bool {
	return s.hasTag("json")
}

// HasYAMLTag returns true if any field in the struct has a yaml tag
func (s *GoStruct) HasYAMLTag() bool {
	return s.hasTag("yaml")
}

// hasTag checks if any field (including nested structs) has the specified tag
func (s *GoStruct) hasTag(tagName string) bool {
	for _, field := range s.Fields {
		if field.Tag != nil && field.Tag.Get(tagName) != "" {
			return true
		}
		// Check nested anonymous structs
		if field.AnonymousStruct != nil && field.AnonymousStruct.hasTag(tagName) {
			return true
		}
	}
	return false
}

// ToJSON generates an example JSON representation of the struct
func (s *GoStruct) ToJSON() string {
	return s.toJSON(0)
}

// ToYAML generates an example YAML representation of the struct
func (s *GoStruct) ToYAML() string {
	return s.toYAML(0)
}

// toJSON is the internal recursive implementation for JSON generation
func (s *GoStruct) toJSON(indent int) string {
	if s == nil || len(s.Fields) == 0 {
		return "{}"
	}

	result := "{\n"
	indentStr := makeIndent(indent + 1)
	fieldCount := 0

	for _, field := range s.Fields {
		if !field.Exported {
			continue
		}

		// Get the JSON field name
		jsonTag := ""
		if field.Tag != nil {
			jsonTag = field.Tag.Get("json")
		}

		// Skip fields with json:"-"
		if jsonTag == "-" {
			continue
		}

		fieldName := getJSONFieldName(field, jsonTag)
		if fieldName == "" {
			continue
		}

		if fieldCount > 0 {
			result += ",\n"
		}

		result += indentStr + `"` + fieldName + `": `
		result += generateJSONValue(field, indent+1)
		fieldCount++
	}

	if fieldCount > 0 {
		result += "\n"
	}
	result += makeIndent(indent) + "}"
	return result
}

// toYAML is the internal recursive implementation for YAML generation
func (s *GoStruct) toYAML(indent int) string {
	if s == nil || len(s.Fields) == 0 {
		return "{}"
	}

	result := ""
	indentStr := makeIndent(indent)
	firstField := true

	for _, field := range s.Fields {
		if !field.Exported {
			continue
		}

		// Get the YAML field name
		yamlTag := ""
		if field.Tag != nil {
			yamlTag = field.Tag.Get("yaml")
		}

		// Skip fields with yaml:"-"
		if yamlTag == "-" {
			continue
		}

		fieldName := getYAMLFieldName(field, yamlTag)
		if fieldName == "" {
			continue
		}

		if !firstField {
			result += "\n"
		}
		firstField = false

		yamlValue := generateYAMLValue(field, indent)
		result += indentStr + fieldName + ":"
		// Don't add space before newline for nested structures
		if !strings.HasPrefix(yamlValue, "\n") {
			result += " "
		}
		result += yamlValue
	}

	return result
}

// GoField is a field in a file or struct
type GoField struct {
	File            *GoFile
	Struct          *GoStruct
	Doc             string
	Decl            string
	Name            string
	Type            string
	Exported        bool
	Tag             *GoTag
	AnonymousStruct *GoStruct
	TypeInfo        *GoType
}

// TypeKind represents the general classification of a Go type expression.
type TypeKind int

const (
	TypeKindUnknown TypeKind = iota
	TypeKindIdent
	TypeKindSelector
	TypeKindPointer
	TypeKindArray
	TypeKindSlice
	TypeKindMap
	TypeKindChan
	TypeKindFunc
	TypeKindStruct
	TypeKindInterface
	TypeKindEllipsis
	TypeKindIndex
	TypeKindIndexList
	TypeKindBinaryExpr
	TypeKindParen
)

// Helper functions for JSON/YAML generation

func makeIndent(level int) string {
	result := ""
	for i := 0; i < level; i++ {
		result += "  "
	}
	return result
}

func getJSONFieldName(field *GoField, jsonTag string) string {
	if jsonTag != "" {
		// Parse json tag (could be "name,omitempty" or just "name")
		parts := strings.Split(jsonTag, ",")
		return parts[0]
	}
	return field.Name
}

func getYAMLFieldName(field *GoField, yamlTag string) string {
	if yamlTag != "" {
		// Parse yaml tag (could be "name,omitempty" or just "name")
		parts := strings.Split(yamlTag, ",")
		return parts[0]
	}
	return field.Name
}

func generateJSONValue(field *GoField, indent int) string {
	// Handle anonymous structs
	if field.AnonymousStruct != nil {
		return field.AnonymousStruct.toJSON(indent)
	}

	// Generate example values based on type
	typeStr := strings.TrimSpace(field.Type)

	// Handle pointers
	if strings.HasPrefix(typeStr, "*") {
		typeStr = strings.TrimPrefix(typeStr, "*")
	}

	// Handle slices and arrays
	if strings.HasPrefix(typeStr, "[]") {
		elemType := strings.TrimPrefix(typeStr, "[]")
		elemValue := generateExampleValueForType(elemType, indent)
		return "[" + elemValue + "]"
	}

	// Handle maps
	if strings.HasPrefix(typeStr, "map[") {
		return `{}`
	}

	return generateExampleValueForType(typeStr, indent)
}

func generateYAMLValue(field *GoField, indent int) string {
	// Handle anonymous structs
	if field.AnonymousStruct != nil {
		yamlContent := field.AnonymousStruct.toYAML(indent + 1)
		if yamlContent == "" || yamlContent == "{}" {
			return "{}"
		}
		return "\n" + yamlContent
	}

	// Generate example values based on type
	typeStr := strings.TrimSpace(field.Type)

	// Handle pointers
	if strings.HasPrefix(typeStr, "*") {
		typeStr = strings.TrimPrefix(typeStr, "*")
	}

	// Handle slices and arrays
	if strings.HasPrefix(typeStr, "[]") {
		elemType := strings.TrimPrefix(typeStr, "[]")
		elemValue := generateExampleValueForType(elemType, indent)
		return "\n" + makeIndent(indent+1) + "- " + elemValue
	}

	// Handle maps
	if strings.HasPrefix(typeStr, "map[") {
		return "{}"
	}

	return generateExampleValueForType(typeStr, indent)
}

func generateExampleValueForType(typeStr string, indent int) string {
	typeStr = strings.TrimSpace(typeStr)

	// Handle basic types
	switch typeStr {
	case "string":
		return `"example"`
	case "int", "int8", "int16", "int32", "int64":
		return "0"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	case "byte":
		return "0"
	case "rune":
		return "0"
	default:
		// For custom types or unknown types
		if strings.Contains(typeStr, ".") {
			// Qualified type from another package
			return `"value"`
		}
		// Assume it's a struct or custom type
		return `{}`
	}
}
