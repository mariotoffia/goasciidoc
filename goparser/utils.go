package goparser

import (
	"go/ast"
	"go/types"
)

func renderTypeName(expr ast.Expr) string {
	return types.ExprString(expr)
}
