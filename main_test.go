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

	defer os.Remove("t.txt")

	arg := args{Overrides: []string{"package=t.txt"}, StdOut: true}

	runner(arg)
}
