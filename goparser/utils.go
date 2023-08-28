package goparser

import (
	"fmt"
	"go/ast"
)

func renderTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return renderTypeName(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + renderTypeName(t.X)
	case *ast.IndexExpr:
		return renderTypeName(t.X) + "[" + renderTypeName(t.Index) + "]"
	default:
		panic(fmt.Sprintf("cannot renderTypeName %T", t))
	}
}
