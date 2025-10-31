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
	File       *GoFile
	Doc        string
	Decl       string
	FullDecl   string
	Name       string
	Exported   bool
	Methods    []*GoMethod
	TypeParams []*GoType
	TypeSet    []*GoType
}

// GoType represents a go type such as a array, map, custom type etc.
type GoType struct {
	File       *GoFile
	Name       string
	Type       string
	Underlying string
	Exported   bool
	Inner      []*GoType
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

// GoField is a field in a file or struct
type GoField struct {
	File     *GoFile
	Struct   *GoStruct
	Doc      string
	Decl     string
	Name     string
	Type     string
	Exported bool
	Tag      *GoTag
	Nested   *GoStruct
}
