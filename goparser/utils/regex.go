package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type regexMatcher struct {
	patterns []compiledRegex
}

type compiledRegex struct {
	raw string
	re  *regexp.Regexp
}

// NewRegexMatcher creates a new regexMatcher from the provided patterns.
// Patterns can be standard regex or glob patterns prefixed with "glb:".
//
// When prefixed with "glb:", it is converted to a regex using `GlobToRegexp`.
func NewRegexMatcher(patterns []string) (*regexMatcher, error) {
	if len(patterns) == 0 {
		return &regexMatcher{}, nil
	}

	compiled := make([]compiledRegex, 0, len(patterns))
	for _, pattern := range patterns {
		p := strings.TrimSpace(pattern)
		if p == "" {
			continue
		}

		var re *regexp.Regexp
		var err error

		if strings.HasPrefix(p, "glb:") {
			re, err = GlobToRegexp(strings.TrimPrefix(p, "glb:"))
			if err != nil {
				return nil, fmt.Errorf("invalid glob exclude pattern %q: %w", pattern, err)
			}
		} else {
			re, err = regexp.Compile(p)
		}

		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}

		compiled = append(compiled, compiledRegex{raw: pattern, re: re})
	}

	return &regexMatcher{patterns: compiled}, nil
}

func (gm *regexMatcher) Match(paths ...string) (bool, string) {
	for _, compiled := range gm.patterns {
		for _, p := range paths {
			if compiled.re.MatchString(p) {
				return true, compiled.raw
			}
		}
	}

	return false, ""
}
