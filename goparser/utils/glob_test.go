package utils_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/mariotoffia/goasciidoc/goparser/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobToRegexpMatches(t *testing.T) {
	tests := []struct {
		name  string
		glob  string
		path  string
		match bool
	}{
		{
			name:  "double star catches nested temp folder",
			glob:  "**/.temp-files/**",
			path:  "/home/user/project/.temp-files/foo/bar.go",
			match: true,
		},
		{
			name:  "double star matches from relative root",
			glob:  "**/.temp-files/**",
			path:  "project/.temp-files/file.go",
			match: true,
		},
		{
			name:  "single star limited to one segment",
			glob:  "src/*/test.go",
			path:  "src/pkg/test.go",
			match: true,
		},
		{
			name:  "single star does not cross extra segment",
			glob:  "src/*/test.go",
			path:  "src/pkg/sub/test.go",
			match: false,
		},
		{
			name:  "alternation matches any listed folder",
			glob:  "**/{bin,obj}/**",
			path:  "dir/obj/lib/file.go",
			match: true,
		},
		{
			name:  "alternation rejects other folder",
			glob:  "**/{bin,obj}/**",
			path:  "dir/tmp/lib/file.go",
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reStr, err := utils.GlobToRegexp(tt.glob)
			require.NoError(t, err)

			re := regexp.MustCompile(reStr)
			got := re.MatchString(filepath.ToSlash(tt.path))
			assert.Equal(t, tt.match, got)
		})
	}
}

func TestGlobToRegexpErrors(t *testing.T) {
	badGlobs := []string{
		"{",
		"**/{a,b",
		"**/{}/**",
		"foo?bar",
		"**/{a,{b,c}}",
	}

	for _, glob := range badGlobs {
		t.Run(glob, func(t *testing.T) {
			_, err := utils.GlobToRegexp(glob)
			require.Error(t, err)
		})
	}
}

func TestRegexMatcherGlbPrefix(t *testing.T) {
	matcher, err := utils.NewRegexMatcher([]string{
		"glb:**/.cache/**",
		`^.*/skip\.go$`,
	})
	require.NoError(t, err)

	match, pattern := matcher.Match("/home/user/.cache/output.txt")
	assert.True(t, match)
	assert.Equal(t, "glb:**/.cache/**", pattern)

	match, pattern = matcher.Match("/home/user/project/skip.go")
	assert.True(t, match)
	assert.Equal(t, `^.*/skip\.go$`, pattern)

	match, pattern = matcher.Match("/home/user/project/main.go")
	assert.False(t, match)
	assert.Empty(t, pattern)
}
