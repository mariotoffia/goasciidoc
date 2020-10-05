package goparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportBaseShallComeFirst(t *testing.T) {
	data := []string{
		"bufio",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/cberror",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/cbevent",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/ctx",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/lambdasupport",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/logger",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/managers/queue",
		"dev.azure.com/dataductus/CbServices/_git/go-core.git/model/deviceshadow",
		"dev.azure.com/dataductus/CbServices/_git/go-services/actuationsvc/actuationcore",
		"encoding/json",
		"flag",
		"fmt",
		"github.com/aws/aws-lambda-go/events",
		"github.com/rs/zerolog",
		"os",
		"strings",
	}

	goFile := &GoFile{}
	for _, s := range data {
		goFile.Imports = append(goFile.Imports, &GoImport{
			Path: s,
		})
	}

	imp := goFile.DeclImports()
	assert.Equal(t, `import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/cberror"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/cbevent"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/ctx"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/lambdasupport"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/logger"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/managers/queue"
	"dev.azure.com/dataductus/CbServices/_git/go-core.git/model/deviceshadow"
	"dev.azure.com/dataductus/CbServices/_git/go-services/actuationsvc/actuationcore"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)`,
		imp)
}
