package processors

import (
	"os"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getPwd() string {
	pwd, _ := os.Getwd()
	return pwd
}

func dummyModule() *goparser.GoModule {

	data := `module github.com/mariotoffia/goasciidoc

	require (
		golang.org/x/mod v0.3.0
		github.com/stretchr/testify v1.6.1
	)
	
	go 1.14`
	path := getPwd() + "go.mod"

	m, err := goparser.NewModuleFromBuff(path, []byte(data))

	if err != nil {
		panic(err)
	}

	return m
}
func TestLinkage(t *testing.T) {

	src := `package foo
	
import "github.com/ahmetb/go-linq/v3"
	
func a() { 
	s := []string{"hej","på","dig"}
	linq.From(s)
}`

	m := dummyModule()
	f, err := goparser.ParseInlineFile(m, m.Base+"/mypkg/file.go", src)
	require.NoError(t, err)
	assert.Equal(t, "foo", f.Package)
	assert.Equal(t, "github.com/ahmetb/go-linq/v3", f.Imports[0].Path)
	assert.Equal(t, "linq", f.Imports[0].Name)
	assert.Equal(t, 0, len(f.Module.Unresolved))

}
