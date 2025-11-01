package goparser

// GoStructMethod is a GoMethod but has receivers and is positioned on a struct or custom type.
type GoStructMethod struct {
	GoMethod
	Receivers     []string
	ReceiverTypes []*GoType
}

// GoMethod is a method on a struct, custom type, interface or just plain function
type GoMethod struct {
	File       *GoFile
	Name       string
	Doc        string
	Decl       string
	FullDecl   string
	Exported   bool
	Params     []*GoType
	Results    []*GoType
	TypeParams []*GoType
}
