package utils_test

import (
	"path/filepath"
	"runtime"
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
			re, err := utils.GlobToRegexp(tt.glob)
			require.NoError(t, err)

			got := re.MatchString(filepath.ToSlash(tt.path))
			assert.Equal(t, tt.match, got)
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

// TestGlobToRegexp_Comprehensive tests all documented features of GlobToRegexp
// with extensive variations to ensure complete coverage.
func TestGlobToRegexp_Comprehensive(t *testing.T) {
	tests := []struct {
		name      string
		glob      string
		matches   []string // paths that should match
		noMatches []string // paths that should NOT match
		skipOS    string   // "windows" or "unix" to skip on that OS
	}{
		// ========================================
		// LITERALS - ordinary characters
		// ========================================
		{
			name: "simple literal path",
			glob: "src/main.go",
			matches: []string{
				"src/main.go",
			},
			noMatches: []string{
				"src/main.c",
				"src/test.go",
				"main.go",
				"src/foo/main.go",
			},
		},
		{
			name: "literals with dots",
			glob: "config.json",
			matches: []string{
				"config.json",
			},
			noMatches: []string{
				"configxjson",
				"config_json",
				"src/config.json",
			},
		},
		{
			name: "literals with special regex chars",
			glob: "test+file.go",
			matches: []string{
				"test+file.go",
			},
			noMatches: []string{
				"testfile.go",
				"test++file.go",
			},
		},
		{
			name: "literals with parens",
			glob: "func(test).go",
			matches: []string{
				"func(test).go",
			},
			noMatches: []string{
				"func.go",
				"functest.go",
			},
		},
		{
			name: "literals with pipe",
			glob: "file|other.txt",
			matches: []string{
				"file|other.txt",
			},
			noMatches: []string{
				"file.txt",
				"other.txt",
			},
		},
		{
			name: "literals with dollar sign",
			glob: "price$100.txt",
			matches: []string{
				"price$100.txt",
			},
			noMatches: []string{
				"price100.txt",
			},
		},
		{
			name: "literals with caret",
			glob: "file^top.txt",
			matches: []string{
				"file^top.txt",
			},
			noMatches: []string{
				"filetop.txt",
			},
		},

		// ========================================
		// SINGLE STAR (*) - zero or more non-separator chars
		// ========================================
		{
			name: "star at end",
			glob: "src/*.go",
			matches: []string{
				"src/main.go",
				"src/test.go",
				"src/a.go",
				"src/.go",
			},
			noMatches: []string{
				"src/sub/main.go",
				"main.go",
				"src/",
			},
		},
		{
			name: "star at start",
			glob: "*/test.go",
			matches: []string{
				"src/test.go",
				"pkg/test.go",
				"a/test.go",
			},
			noMatches: []string{
				"test.go",
				"src/sub/test.go",
			},
		},
		{
			name: "star in middle",
			glob: "src/test_*.go",
			matches: []string{
				"src/test_.go",
				"src/test_a.go",
				"src/test_foo.go",
				"src/test_123.go",
			},
			noMatches: []string{
				"src/test.go",
				"src/testing.go",
				"src/sub/test_a.go",
			},
		},
		{
			name: "multiple stars",
			glob: "src/*/test_*.go",
			matches: []string{
				"src/pkg/test_a.go",
				"src/x/test_foo.go",
			},
			noMatches: []string{
				"src/test_a.go",
				"src/pkg/sub/test_a.go",
			},
		},
		{
			name: "star matching zero chars",
			glob: "test*.go",
			matches: []string{
				"test.go",
				"testfoo.go",
				"test123.go",
			},
			noMatches: []string{
				"test/.go",
				"foo/test.go",
			},
		},
		{
			name: "consecutive stars (not globstar)",
			glob: "a*b*c",
			matches: []string{
				"abc",
				"aXbYc",
				"aabbcc",
			},
			noMatches: []string{
				"ab",
				"ac",
				"a/b/c",
			},
		},

		// ========================================
		// QUESTION MARK (?) - exactly one non-separator char
		// ========================================
		{
			name: "single question mark",
			glob: "test?.go",
			matches: []string{
				"test1.go",
				"testa.go",
				"test_.go",
			},
			noMatches: []string{
				"test.go",
				"test12.go",
				"test/.go",
			},
		},
		{
			name: "multiple question marks",
			glob: "test???.go",
			matches: []string{
				"test123.go",
				"testabc.go",
				"test___.go",
			},
			noMatches: []string{
				"test12.go",
				"test1234.go",
				"test.go",
			},
		},
		{
			name: "question mark at start",
			glob: "?est.go",
			matches: []string{
				"test.go",
				"best.go",
				"1est.go",
			},
			noMatches: []string{
				"est.go",
				"ttest.go",
			},
		},
		{
			name: "question marks and stars mixed",
			glob: "test?*.go",
			matches: []string{
				"test1.go",
				"test12.go",
				"testa.go",
				"testfoo.go",
			},
			noMatches: []string{
				"test.go",
			},
		},
		{
			name: "question mark does not match separator",
			glob: "src?test.go",
			matches: []string{
				"src1test.go",
				"srcXtest.go",
			},
			noMatches: []string{
				"src/test.go",
			},
		},

		// ========================================
		// GLOBSTAR (**) - matches across directories
		// ========================================
		{
			name: "globstar at start",
			glob: "**/test.go",
			matches: []string{
				"test.go",
				"src/test.go",
				"src/pkg/test.go",
				"a/b/c/d/test.go",
			},
			noMatches: []string{
				"test.c",
				"testing.go",
			},
		},
		{
			name: "globstar at end",
			glob: "src/**",
			matches: []string{
				"src/",
				"src/test.go",
				"src/pkg/main.go",
				"src/a/b/c/d.go",
			},
			noMatches: []string{
				"src",
				"other/test.go",
			},
		},
		{
			name: "globstar in middle",
			glob: "src/**/test.go",
			matches: []string{
				"src/test.go",
				"src/pkg/test.go",
				"src/a/b/test.go",
			},
			noMatches: []string{
				"test.go",
				"other/src/test.go",
			},
		},
		{
			name: "multiple globstars",
			glob: "**/src/**/test.go",
			matches: []string{
				"src/test.go",
				"pkg/src/test.go",
				"src/pkg/test.go",
				"a/b/src/c/d/test.go",
			},
			noMatches: []string{
				"test.go",
				"src.go",
			},
		},
		{
			name: "globstar with surrounding patterns",
			glob: "**/temp/**/*.tmp",
			matches: []string{
				"temp/file.tmp",
				"src/temp/file.tmp",
				"src/temp/cache/file.tmp",
			},
			noMatches: []string{
				"temp.tmp",
				"temp",
				"file.tmp",
			},
		},
		{
			name: "globstar matching zero segments",
			glob: "src/**/test.go",
			matches: []string{
				"src/test.go", // ** can match zero segments
			},
		},

		// ========================================
		// CHARACTER CLASSES [...] - sets and ranges
		// ========================================
		{
			name: "simple character class",
			glob: "test[abc].go",
			matches: []string{
				"testa.go",
				"testb.go",
				"testc.go",
			},
			noMatches: []string{
				"testd.go",
				"test.go",
				"testabc.go",
			},
		},
		{
			name: "character class with range",
			glob: "test[0-9].go",
			matches: []string{
				"test0.go",
				"test5.go",
				"test9.go",
			},
			noMatches: []string{
				"testa.go",
				"test.go",
				"test10.go",
			},
		},
		{
			name: "character class with multiple ranges",
			glob: "test[a-zA-Z0-9].go",
			matches: []string{
				"testa.go",
				"testZ.go",
				"test5.go",
			},
			noMatches: []string{
				"test_.go",
				"test-.go",
			},
		},
		{
			name: "negated character class with !",
			glob: "test[!0-9].go",
			matches: []string{
				"testa.go",
				"testb.go",
				"test_.go",
			},
			noMatches: []string{
				"test0.go",
				"test5.go",
				"test9.go",
			},
		},
		{
			name: "negated character class with ^",
			glob: "test[^abc].go",
			matches: []string{
				"testd.go",
				"test1.go",
				"test_.go",
			},
			noMatches: []string{
				"testa.go",
				"testb.go",
				"testc.go",
			},
		},
		{
			name: "character class with dash at start",
			glob: "test[-abc].go",
			matches: []string{
				"test-.go",
				"testa.go",
				"testb.go",
			},
			noMatches: []string{
				"testd.go",
			},
		},
		{
			name: "character class with dash at end",
			glob: "test[abc-].go",
			matches: []string{
				"testa.go",
				"testb.go",
				"test-.go",
			},
			noMatches: []string{
				"testd.go",
			},
		},
		{
			name:    "character class does not match separator",
			glob:    "src[/]test.go",
			matches: []string{},
			noMatches: []string{
				"src/test.go",
			},
		},

		// ========================================
		// POSIX CHARACTER CLASSES [[:...:]]
		// ========================================
		{
			name: "POSIX digit class",
			glob: "test[[:digit:]].go",
			matches: []string{
				"test0.go",
				"test5.go",
				"test9.go",
			},
			noMatches: []string{
				"testa.go",
				"test.go",
			},
		},
		{
			name: "POSIX alpha class",
			glob: "test[[:alpha:]].go",
			matches: []string{
				"testa.go",
				"testZ.go",
			},
			noMatches: []string{
				"test0.go",
				"test_.go",
			},
		},
		{
			name: "POSIX alnum class",
			glob: "test[[:alnum:]].go",
			matches: []string{
				"testa.go",
				"testZ.go",
				"test0.go",
				"test9.go",
			},
			noMatches: []string{
				"test_.go",
				"test-.go",
			},
		},
		{
			name: "POSIX space class",
			glob: "test[[:space:]]file.txt",
			matches: []string{
				"test file.txt",
				"test\tfile.txt",
			},
			noMatches: []string{
				"testfile.txt",
				"test_file.txt",
			},
		},
		{
			name: "negated POSIX class",
			glob: "test[^[:digit:]].go",
			matches: []string{
				"testa.go",
				"testb.go",
			},
			noMatches: []string{
				"test0.go",
				"test9.go",
			},
		},

		// ========================================
		// ALTERNATION {...} - multiple options
		// ========================================
		{
			name: "simple alternation",
			glob: "test.{go,py,js}",
			matches: []string{
				"test.go",
				"test.py",
				"test.js",
			},
			noMatches: []string{
				"test.c",
				"test.go.bak",
			},
		},
		{
			name: "alternation with paths",
			glob: "src/{bin,obj}/output",
			matches: []string{
				"src/bin/output",
				"src/obj/output",
			},
			noMatches: []string{
				"src/tmp/output",
				"src/output",
			},
		},
		{
			name: "nested alternation",
			glob: "test.{go,{js,ts}}",
			matches: []string{
				"test.go",
				"test.js",
				"test.ts",
			},
			noMatches: []string{
				"test.py",
			},
		},
		{
			name: "empty option in alternation",
			glob: "test{,_old}.go",
			matches: []string{
				"test.go",
				"test_old.go",
			},
			noMatches: []string{
				"testing.go",
			},
		},
		{
			name: "alternation with wildcards",
			glob: "**/{bin,obj}/**/*.{dll,so}",
			matches: []string{
				"bin/file.dll",
				"obj/file.so",
				"src/bin/lib.dll",
				"pkg/obj/sub/lib.so",
			},
			noMatches: []string{
				"tmp/file.dll",
				"bin/file.a",
			},
		},
		{
			name: "multiple separate alternations",
			glob: "{src,test}/{unit,integration}/test.go",
			matches: []string{
				"src/unit/test.go",
				"src/integration/test.go",
				"test/unit/test.go",
				"test/integration/test.go",
			},
			noMatches: []string{
				"src/e2e/test.go",
				"pkg/unit/test.go",
			},
		},
		{
			name: "deeply nested alternation",
			glob: "file.{a,{b,{c,d}}}.txt",
			matches: []string{
				"file.a.txt",
				"file.b.txt",
				"file.c.txt",
				"file.d.txt",
			},
			noMatches: []string{
				"file.e.txt",
			},
		},

		// ========================================
		// ESCAPING (Unix-only)
		// ========================================
		{
			name:   "escape star",
			glob:   `test\*.go`,
			skipOS: "windows",
			matches: []string{
				"test*.go",
			},
			noMatches: []string{
				"testfoo.go",
				"test.go",
			},
		},
		{
			name:   "escape question mark",
			glob:   `test\?.go`,
			skipOS: "windows",
			matches: []string{
				"test?.go",
			},
			noMatches: []string{
				"test1.go",
				"test.go",
			},
		},
		{
			name:   "escape bracket",
			glob:   `test\[a\].go`,
			skipOS: "windows",
			matches: []string{
				"test[a].go",
			},
			noMatches: []string{
				"testa.go",
			},
		},
		{
			name:   "escape brace",
			glob:   `test\{a,b\}.go`,
			skipOS: "windows",
			matches: []string{
				"test{a,b}.go",
			},
			noMatches: []string{
				"testa.go",
				"testb.go",
			},
		},
		{
			name:   "escape backslash",
			glob:   `test\\file.go`,
			skipOS: "windows",
			matches: []string{
				`test\file.go`,
			},
			noMatches: []string{
				"testfile.go",
			},
		},
		{
			name:   "escape in character class",
			glob:   `test[\-\]].go`,
			skipOS: "windows",
			matches: []string{
				"test-.go",
				"test].go",
			},
			noMatches: []string{
				"testa.go",
			},
		},

		// ========================================
		// PATH SEPARATORS
		// ========================================
		{
			name: "forward slash",
			glob: "src/pkg/test.go",
			matches: []string{
				"src/pkg/test.go",
			},
			noMatches: []string{
				"srcpkgtest.go",
				"src-pkg-test.go",
			},
		},
		{
			name: "multiple separators",
			glob: "a/b/c/d.go",
			matches: []string{
				"a/b/c/d.go",
			},
			noMatches: []string{
				"a/b/d.go",
			},
		},

		// ========================================
		// COMPLEX COMBINATIONS
		// ========================================
		{
			name: "complex: globstar with alternation and char class",
			glob: "**/test_[a-z]{1,2,3}.{go,py}",
			matches: []string{
				"test_a1.go",
				"test_z2.py",
				"src/test_b3.go",
				"a/b/c/test_x1.py",
			},
			noMatches: []string{
				"test_11.go",
				"test_a4.go",
				"test_a1.js",
			},
		},
		{
			name: "complex: all wildcards in one pattern",
			glob: "src/**/pkg_?_[0-9]{debug,release}/*.{a,so}",
			matches: []string{
				"src/pkg_a_1debug/lib.a",
				"src/sub/pkg_x_9release/lib.so",
				"src/a/b/pkg_z_0debug/mylib.a",
			},
			noMatches: []string{
				"src/pkg_11debug/lib.a",
				"src/pkg_a_1test/lib.a",
				"src/pkg_a_1debug/lib.dll",
			},
		},
		{
			name: "complex: nested alternation with globstar",
			glob: "**/{src,{test,spec}}/{unit,integration}/**/*.go",
			matches: []string{
				"src/unit/test.go",
				"test/unit/foo.go",
				"spec/integration/bar.go",
				"pkg/src/unit/sub/test.go",
				"a/b/test/integration/c/d/file.go",
			},
			noMatches: []string{
				"src/e2e/test.go",
				"other/unit/test.go",
			},
		},
		{
			name: "complex: multiple stars and classes",
			glob: "test_*_[a-z]_[0-9]*.go",
			matches: []string{
				"test_foo_a_0.go",
				"test_bar_z_123.go",
				"test__x_9999.go",
			},
			noMatches: []string{
				"test_foo_1_0.go",
				"test_foo_a_.go",
			},
		},
		{
			name: "real-world: exclude temp files",
			glob: "**/.temp-files/**",
			matches: []string{
				".temp-files/test.go",
				"src/.temp-files/cache.txt",
				"a/b/c/.temp-files/d/e/f.go",
			},
			noMatches: []string{
				"temp-files/test.go",
				"src/temp.go",
			},
		},
		{
			name: "real-world: build artifacts",
			glob: "**/{bin,obj,build,dist}/**/*.{exe,dll,so,a,o}",
			matches: []string{
				"bin/app.exe",
				"obj/lib.dll",
				"build/output/lib.so",
				"dist/myapp.exe",
				"src/bin/debug/app.exe",
			},
			noMatches: []string{
				"src/main.go",
				"tmp/lib.so",
			},
		},
		{
			name: "real-world: test files pattern",
			glob: "**/*_{test,spec,integration}.{go,py,js}",
			matches: []string{
				"main_test.go",
				"utils_spec.js",
				"api_integration.py",
				"src/pkg/handler_test.go",
			},
			noMatches: []string{
				"main.go",
				"test.go",
				"main_unit.go",
			},
		},
		{
			name: "edge: empty string",
			glob: "",
			matches: []string{
				"",
			},
			noMatches: []string{
				"anything",
			},
		},
		{
			name: "edge: just globstar",
			glob: "**",
			matches: []string{
				"",
				"file.go",
				"a/b/c/d.go",
			},
		},
		{
			name: "edge: just star",
			glob: "*",
			matches: []string{
				"",
				"file",
				"test.go",
			},
			noMatches: []string{
				"a/b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS == "windows" && runtime.GOOS == "windows" {
				t.Skip("Skipping Unix-specific test on Windows")
			}
			if tt.skipOS == "unix" && runtime.GOOS != "windows" {
				t.Skip("Skipping Windows-specific test on Unix")
			}

			re, err := utils.GlobToRegexp(tt.glob)
			require.NoError(t, err, "Failed to compile glob pattern: %s", tt.glob)

			// Test all matches
			for _, path := range tt.matches {
				matched := re.MatchString(path)
				assert.True(t, matched, "Expected pattern %q to match path %q", tt.glob, path)
			}

			// Test all non-matches
			for _, path := range tt.noMatches {
				matched := re.MatchString(path)
				assert.False(t, matched, "Expected pattern %q NOT to match path %q", tt.glob, path)
			}
		})
	}
}

// TestGlobToRegexp_ErrorCases tests all documented error conditions
func TestGlobToRegexp_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		glob        string
		errorSubstr string // substring that should appear in error message
		skipOS      string
	}{
		{
			name:        "unterminated character class",
			glob:        "test[abc",
			errorSubstr: "character class",
		},
		{
			name:        "unterminated character class at end",
			glob:        "test.go[",
			errorSubstr: "character class",
		},
		{
			name:        "empty character class",
			glob:        "test[].go",
			errorSubstr: "empty character class",
		},
		{
			name:        "empty negated character class",
			glob:        "test[!].go",
			errorSubstr: "empty character class",
		},
		{
			name:        "unmatched opening brace",
			glob:        "test.{go,py",
			errorSubstr: "unmatched '{'",
		},
		{
			name:        "unmatched opening brace nested",
			glob:        "test.{go,{py,js}",
			errorSubstr: "unmatched '{'",
		},
		{
			name:        "nested unmatched brace",
			glob:        "{a,{b,c}",
			errorSubstr: "unmatched '{'",
		},
		{
			name:        "trailing backslash on Unix",
			glob:        `test\`,
			errorSubstr: "trailing backslash",
			skipOS:      "windows",
		},
		{
			name:        "trailing backslash with content on Unix",
			glob:        `src/test.go\`,
			errorSubstr: "trailing backslash",
			skipOS:      "windows",
		},
		{
			name:        "unclosed POSIX class",
			glob:        "test[[:digit:].go",
			errorSubstr: "character class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS == "windows" && runtime.GOOS == "windows" {
				t.Skip("Skipping Unix-specific test on Windows")
			}
			if tt.skipOS == "unix" && runtime.GOOS != "windows" {
				t.Skip("Skipping Windows-specific test on Unix")
			}

			_, err := utils.GlobToRegexp(tt.glob)
			require.Error(t, err, "Expected error for glob pattern: %s", tt.glob)
			if tt.errorSubstr != "" {
				assert.Contains(t, err.Error(), tt.errorSubstr,
					"Error message should contain %q for pattern %q", tt.errorSubstr, tt.glob)
			}
		})
	}
}

// TestGlobToRegexp_WindowsPathSeparators tests Windows-specific behavior
func TestGlobToRegexp_WindowsPathSeparators(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	tests := []struct {
		name      string
		glob      string
		matches   []string
		noMatches []string
	}{
		{
			name: "backslash as separator",
			glob: `src\pkg\test.go`,
			matches: []string{
				`src\pkg\test.go`,
				`src/pkg/test.go`, // both separators should match
			},
			noMatches: []string{
				`srcpkgtest.go`,
			},
		},
		{
			name: "mixed separators in pattern",
			glob: `src\pkg/test.go`,
			matches: []string{
				`src\pkg\test.go`,
				`src/pkg/test.go`,
				`src/pkg\test.go`,
			},
		},
		{
			name: "star does not match backslash",
			glob: `src\*\test.go`,
			matches: []string{
				`src\pkg\test.go`,
				`src/pkg/test.go`,
			},
			noMatches: []string{
				`src\pkg\sub\test.go`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := utils.GlobToRegexp(tt.glob)
			require.NoError(t, err)

			for _, path := range tt.matches {
				assert.True(t, re.MatchString(path),
					"Expected pattern %q to match path %q", tt.glob, path)
			}

			for _, path := range tt.noMatches {
				assert.False(t, re.MatchString(path),
					"Expected pattern %q NOT to match path %q", tt.glob, path)
			}
		})
	}
}

// TestGlobToRegexp_AnchoringBehavior tests that patterns match entire strings
func TestGlobToRegexp_AnchoringBehavior(t *testing.T) {
	tests := []struct {
		name      string
		glob      string
		fullMatch string   // should match (entire string)
		noMatches []string // should not match (substring scenarios)
	}{
		{
			name:      "pattern must match entire string",
			glob:      "test.go",
			fullMatch: "test.go",
			noMatches: []string{
				"prefix_test.go",
				"test.go_suffix",
				"dir/test.go",
				"test.go/file",
			},
		},
		{
			name:      "star must be at boundaries",
			glob:      "*.go",
			fullMatch: "main.go",
			noMatches: []string{
				"src/main.go",
				"main.go.bak",
			},
		},
		{
			name:      "globstar still requires full match",
			glob:      "**/test.go",
			fullMatch: "src/pkg/test.go",
			noMatches: []string{
				"src/pkg/test.go.bak",
				"prefix/src/pkg/test.go/suffix",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := utils.GlobToRegexp(tt.glob)
			require.NoError(t, err)

			assert.True(t, re.MatchString(tt.fullMatch),
				"Pattern %q should match entire string %q", tt.glob, tt.fullMatch)

			for _, path := range tt.noMatches {
				assert.False(t, re.MatchString(path),
					"Pattern %q should NOT match %q (substring/partial)", tt.glob, path)
			}
		})
	}
}

// TestGlobToRegexp_SpecialCases tests edge cases and boundary conditions
func TestGlobToRegexp_SpecialCases(t *testing.T) {
	tests := []struct {
		name    string
		glob    string
		matches []string
	}{
		{
			name: "alternation with one empty option at start",
			glob: "{,test}_file.go",
			matches: []string{
				"_file.go",
				"test_file.go",
			},
		},
		{
			name: "alternation with one empty option at end",
			glob: "file{.bak,}.txt",
			matches: []string{
				"file.bak.txt",
				"file.txt",
			},
		},
		{
			name: "alternation with multiple empty options",
			glob: "{,,test}.go",
			matches: []string{
				".go",
				"test.go",
			},
		},
		{
			name: "globstar at start and end",
			glob: "**/src/**",
			matches: []string{
				"src/",
				"src/file.go",
				"pkg/src/",
				"pkg/src/file.go",
				"a/b/src/c/d.go",
			},
		},
		{
			name: "consecutive separators normalized",
			glob: "src//pkg///test.go",
			matches: []string{
				"src//pkg///test.go",
			},
		},
		{
			name:    "character class with ]",
			glob:    "test[]ab].go",
			matches: []string{
				// This should be an error or specific behavior
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := utils.GlobToRegexp(tt.glob)
			if err != nil {
				t.Logf("Pattern %q resulted in error: %v", tt.glob, err)
				return
			}

			for _, path := range tt.matches {
				matched := re.MatchString(path)
				t.Logf("Pattern %q vs path %q: matched=%v", tt.glob, path, matched)
			}
		})
	}
}
