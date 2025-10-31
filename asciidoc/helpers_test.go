package asciidoc

import (
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
)

func TestTypeSetItemsDeduplicatesAndTrims(t *testing.T) {
	raw := []*goparser.GoType{
		{Type: "Reader"},
		{Type: "Reader"},
		{Type: "(Writer)"},
		{Type: "  ~[]T  "},
		nil,
	}

	items := typeSetItems(raw)
	assert.ElementsMatch(t, []string{"Reader", "Writer", "~[]T"}, items)
}
