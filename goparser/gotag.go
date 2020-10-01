package goparser

import (
	"reflect"
	"strings"
)

// GoTag is a tag on a struct field
type GoTag struct {
	File  *GoFile
	Field *GoField
	Value string
}

// Get returns a struct tag with the specified name e.g. json
func (g *GoTag) Get(key string) string {
	tag := strings.Replace(g.Value, "`", "", -1)
	return reflect.StructTag(tag).Get(key)
}
