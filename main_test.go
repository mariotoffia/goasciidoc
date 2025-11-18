package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mariotoffia/goasciidoc/asciidoc"
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

func TestParseHighlighter(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"", "highlightjs"},
		{"none", "none"},
		{"goasciidoc", "goasciidoc"},
		{"highlight", "highlightjs"},
		{"highlightjs", "highlightjs"},
		{"highlight.js", "highlightjs"},
	}

	for _, tc := range tests {
		got, err := parseHighlighter(tc.input)
		assert.NoError(t, err, tc.input)
		assert.Equal(t, tc.expect, got)
	}

	_, err := parseHighlighter("unknown")
	assert.Error(t, err)
}

func TestParsePackageMode(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expect     string
		shouldFail bool
	}{
		{
			name:       "empty defaults to none",
			input:      "",
			expect:     "none",
			shouldFail: false,
		},
		{
			name:       "explicit none",
			input:      "none",
			expect:     "none",
			shouldFail: false,
		},
		{
			name:       "include mode",
			input:      "include",
			expect:     "include",
			shouldFail: false,
		},
		{
			name:       "link mode",
			input:      "link",
			expect:     "link",
			shouldFail: false,
		},
		{
			name:       "case insensitive - Include",
			input:      "Include",
			expect:     "include",
			shouldFail: false,
		},
		{
			name:       "case insensitive - LINK",
			input:      "LINK",
			expect:     "link",
			shouldFail: false,
		},
		{
			name:       "case insensitive - NONE",
			input:      "NONE",
			expect:     "none",
			shouldFail: false,
		},
		{
			name:       "whitespace trimmed",
			input:      "  include  ",
			expect:     "include",
			shouldFail: false,
		},
		{
			name:       "invalid mode",
			input:      "invalid",
			expect:     "",
			shouldFail: true,
		},
		{
			name:       "typo - separated",
			input:      "separated",
			expect:     "",
			shouldFail: true,
		},
		{
			name:       "typo - single",
			input:      "single",
			expect:     "",
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePackageMode(tc.input)

			if tc.shouldFail {
				assert.Error(t, err, "expected error for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "unexpected error for input: %s", tc.input)
				// Verify the mode value
				var expectedMode asciidoc.PackageMode
				switch tc.expect {
				case "none":
					expectedMode = asciidoc.PackageModeNone
				case "include":
					expectedMode = asciidoc.PackageModeInclude
				case "link":
					expectedMode = asciidoc.PackageModeLink
				}
				assert.Equal(t, expectedMode, got, "mode mismatch for input: %s", tc.input)
			}
		})
	}
}

func TestParsePackageModeImport(t *testing.T) {
	// Import test to ensure package mode types are accessible
	_ = asciidoc.PackageModeNone
	_ = asciidoc.PackageModeInclude
	_ = asciidoc.PackageModeLink
}
