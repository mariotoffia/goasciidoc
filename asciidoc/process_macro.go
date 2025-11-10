package asciidoc

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mariotoffia/goasciidoc/goparser"
)

var (
	// asciidocTagRegex cleans out asciidoc tags from the source
	asciidocTagRegex = regexp.MustCompile(`\s*tag::.*\[\]`)

	// mdHeadingRegex matches markdown headings at the start of a line
	// Matches: # text, ## text, ### text, etc.
	// Also handles optional leading whitespace and optional text after heading markers
	mdHeadingRegex = regexp.MustCompile(`(?m)^[ \t]*(#{1,6})(?:[ \t]+.*)?$`)

	// mdHeadingExtractRegex extracts the text content after heading markers
	mdHeadingExtractRegex = regexp.MustCompile(`^[ \t]*#{1,6}[ \t]*(.*)$`)
)

// getProcessMacroFunc returns the macro substitution function.
func getProcessMacroFunc(
	config goparser.ParseConfig,
	next goparser.ParseSinglePackageWalkerFunc) goparser.ParseSinglePackageWalkerFunc {

	return func(pkg *goparser.GoPackage) error {

		processDocs := func(doc, fq string) string {

			fqdir := filepath.Dir(fq)
			dir := filepath.Base(fqdir)
			file := filepath.Base(fq)

			doc = strings.ReplaceAll(doc, "${gad:current:fq}", fq)
			doc = strings.ReplaceAll(doc, "${gad:current:fqdir}", fqdir)
			doc = strings.ReplaceAll(doc, "${gad:current:dir}", dir)
			doc = strings.ReplaceAll(doc, "${gad:current:file}", file)

			doc = asciidocTagRegex.ReplaceAllString(doc, "")

			// If IgnoreMarkdownHeadings is enabled, replace markdown headings with their text content
			if config.IgnoreMarkdownHeadings {
				// Remove leading whitespace and heading markers, keeping only the text if any
				doc = mdHeadingRegex.ReplaceAllStringFunc(doc, func(match string) string {
					// Extract the text after the heading markers
					if matches := mdHeadingExtractRegex.FindStringSubmatch(match); len(matches) > 1 {
						return strings.TrimRight(matches[1], " \t")
					}
					return ""
				})
			}

			return doc
		}

		pkg.Doc = processDocs(pkg.Doc, pkg.Module.Base)

		for _, c := range pkg.ConstAssignments {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.CustomFuncs {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.CustomTypes {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Imports {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Interfaces {
			c.Doc = processDocs(c.Doc, c.File.FilePath)

			for _, c := range c.Methods {
				c.Doc = processDocs(c.Doc, c.File.FilePath)
			}

		}

		for _, c := range pkg.StructMethods {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		for _, c := range pkg.Structs {
			c.Doc = processDocs(c.Doc, c.File.FilePath)

			for _, c := range c.Fields {
				c.Doc = processDocs(c.Doc, c.File.FilePath)
			}

		}

		for _, c := range pkg.VarAssignments {
			c.Doc = processDocs(c.Doc, c.File.FilePath)
		}

		return next(pkg)
	}

}
