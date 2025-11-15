// Package goparser was taken from an open source project (https://github.com/zpatrick/go-parser) by zpatrick. Since it seemed
// that he had abandon it, I've integrated it into this project (and extended it).
package goparser

import (
	"go/ast"
	"go/token"
	"strings"
)

// extractDocs extracts documentation text from a comment group.
func extractDocs(doc *ast.CommentGroup) string {
	d := doc.Text()
	if d == "" {
		return d
	}

	return d[:len(d)-1]
}

// docString extracts and optionally concatenates documentation from comment groups.
// It supports both DocConcatenationNone and DocConcatenationFull modes.
func docString(ctx *parseContext, doc *ast.CommentGroup, declPos token.Pos) string {
	if doc == nil {
		return ""
	}
	base := extractDocs(doc)
	if ctx == nil || ctx.docCtx == nil || ctx.docCtx.mode != DocConcatenationFull {
		return base
	}
	comments := ctx.docCtx.comments
	if len(comments) == 0 {
		return base
	}
	index := -1
	for i, group := range comments {
		if group == doc {
			index = i
			break
		}
	}
	if index == -1 {
		return base
	}

	type segment struct {
		text    string
		between string
	}

	var builder strings.Builder

	cursorStart := doc.Pos()
	preceding := []segment{}
	for i := index - 1; i >= 0; i-- {
		group := comments[i]
		if group == nil || !group.End().IsValid() {
			continue
		}
		between := ctx.docCtx.src.slice(group.End(), cursorStart)
		if strings.TrimSpace(between) != "" {
			break
		}
		text := extractDocs(group)
		if text == "" {
			cursorStart = group.Pos()
			continue
		}
		preceding = append(preceding, segment{text: text, between: between})
		cursorStart = group.Pos()
	}
	for i := len(preceding) - 1; i >= 0; i-- {
		builder.WriteString(preceding[i].text)
		builder.WriteString(preceding[i].between)
	}

	builder.WriteString(base)
	cursor := doc.End()

	for _, group := range comments[index+1:] {
		if group == nil || !group.Pos().IsValid() {
			continue
		}
		if declPos.IsValid() && group.Pos() >= declPos {
			break
		}
		between := ctx.docCtx.src.slice(cursor, group.Pos())
		if strings.TrimSpace(between) != "" {
			break
		}
		text := extractDocs(group)
		if text == "" {
			cursor = group.End()
			continue
		}
		builder.WriteString(between)
		builder.WriteString(text)
		cursor = group.End()
	}

	return builder.String()
}

// extractBuildTagsFromComments extracts build tags from file comments.
// It supports both //go:build (newer) and // +build (older) directives.
func extractBuildTagsFromComments(comments []*ast.CommentGroup) []string {
	if len(comments) == 0 {
		return nil
	}

	var tags []string
	seen := make(map[string]bool)

	for _, cg := range comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(c.Text)

			// Check for //go:build directive (newer style)
			if strings.HasPrefix(text, "//go:build ") {
				constraint := strings.TrimPrefix(text, "//go:build ")
				constraint = strings.TrimSpace(constraint)
				if constraint != "" && !seen[constraint] {
					tags = append(tags, constraint)
					seen[constraint] = true
				}
			}

			// Check for // +build directive (older style)
			if strings.HasPrefix(text, "// +build ") {
				constraint := strings.TrimPrefix(text, "// +build ")
				constraint = strings.TrimSpace(constraint)
				if constraint != "" && !seen[constraint] {
					tags = append(tags, constraint)
					seen[constraint] = true
				}
			}
		}
	}

	return tags
}
