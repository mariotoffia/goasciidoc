package asciidoc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mariotoffia/goasciidoc/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateWorkspaceToString(t *testing.T) {
	target, err := utils.TempCopyDir("../", "asciidoc-tests")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer os.RemoveAll(target)

	var buf bytes.Buffer

	p := NewProducer().
		Include(target).
		Writer(&buf).
		Outfile(filepath.Join(target, "_docs", "package.adoc")).
		Module(target)

	p.Generate()

	fmt.Println(buf.String())
}
