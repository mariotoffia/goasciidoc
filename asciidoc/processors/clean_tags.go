package processors

import (
	"regexp"

	"github.com/mariotoffia/goasciidoc/goparser"
)

// CleanTagsProcessor will clean all _tag::tag-name[]_ of all
// locations where begins a line with the _tag_ directive.
func CleanTagsProcessor() DocProcessor {

	// Clean out asciidoc tags from the source
	r, _ := regexp.Compile(`\s*tag::.*\[\]`)

	return func(module *goparser.GoModule, file *goparser.GoFile, doc, fq string) string {

		return r.ReplaceAllString(doc, "")

	}
}
