package goparser

// GoAssignment represents a single var assignment e.g. var pelle = 10
type GoAssignment struct {
	File *GoFile
	Name string
	Doc  string
	// Decl will be the same if multi var assignment on same row e.g. var pelle, lisa = 10, 19
	// then both pelle and list will have 'var pelle, lisa = 10, 19' as Decl
	Decl     string
	FullDecl string
}

// GoCustomType is a custom type definition
type GoCustomType struct {
	File *GoFile
	Name string
	Doc  string
	Type string
	Decl string
}

// GoInterface specifies a interface definition
type GoInterface struct {
	File     *GoFile
	Doc      string
	Decl     string
	FullDecl string
	Name     string
	Methods  []*GoMethod
}

// GoMethod is a method on a struct, interface or just plain function
type GoMethod struct {
	File     *GoFile
	Name     string
	Doc      string
	Decl     string
	FullDecl string
	Params   []*GoType
	Results  []*GoType
}

// GoStructMethod is a GoMethod but has receivers and is positioned on a struct.
type GoStructMethod struct {
	GoMethod
	Receivers []string
}

// GoType represents a go type such as a array, map, custom type etc.
type GoType struct {
	File       *GoFile
	Name       string
	Type       string
	Underlying string
	Inner      []*GoType
}

// GoStruct represents a struct
type GoStruct struct {
	File     *GoFile
	Doc      string
	Decl     string
	FullDecl string
	Name     string
	Fields   []*GoField
}

// GoField is a field in a file or struct
type GoField struct {
	File   *GoFile
	Struct *GoStruct
	Doc    string
	Decl   string
	Name   string
	Type   string
	Tag    *GoTag
	Nested *GoStruct
}
