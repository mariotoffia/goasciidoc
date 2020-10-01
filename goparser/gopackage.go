package goparser

// GoPackage is a aggregation of all GoFiles in a single
// package for ease of access.
type GoPackage struct {
	GoFile
	// Files are all files in current package.
	Files []*GoFile
}
