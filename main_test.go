package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverridePackageTemplate(t *testing.T) {

	if err := ioutil.WriteFile("t.txt",
		[]byte(`== Override Package {{.File.FqPackage}}`),
		0644,
	); err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		os.Remove("t.txt")
		os.Remove("test-docs.adoc")
	}()
	arg := args{Overrides: []string{"package=t.txt"}, Out: "test-docs.adoc"}

	runner(arg)
}

func TestTemplateDir(t *testing.T) {

	defer func() {
		os.Remove("test-docs.adoc")
	}()

	arg := args{TemplateDir: "defaults", Out: "test-docs.adoc"}

	runner(arg)
}

func TestNonExported(t *testing.T) {

	defer os.Remove("test-docs.adoc")
	arg := args{NonExported: true, Out: "test-docs.adoc"}

	runner(arg)
}
