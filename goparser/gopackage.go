package goparser

// GoPackage is a aggregation of all GoFiles in a single
// package for ease of access.
type GoPackage struct {
	// Module is where this package belongs to.
	Module *GoModule
	// Files are all files in current package.
	Files []*GoFile
	// Package name (short name)
	Package string
	// Path is the fully qualified path to the package directory.
	Path string
	// Doc is the package documentation
	Doc string
	// Decl is the package declaration
	Decl string
	// Structs is all the structs in the package
	Structs []*GoStruct
	// Interfaces are all the interfaces in the package
	Interfaces []*GoInterface
	// Imports are all the imports combined in the package
	Imports []*GoImport
	// StructMethods is all struct methods in the system
	StructMethods []*GoStructMethod
	// CustomTypes are all custom variable type definition in the package
	CustomTypes []*GoCustomType
	// CustomFuncs are all custom typedefs for functions in the package
	CustomFuncs []*GoMethod
	// VarAssigments is all var assignments in the package
	VarAssigments []*GoAssignment
	// ConstAssignments is all const assignments in the package
	ConstAssignments []*GoAssignment
}
