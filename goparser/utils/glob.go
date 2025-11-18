package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// GlobToRegexp converts a limited glob pattern to a regexp string.
// Supported tokens: ** (across path separators), * (within a path segment), {a,b,c} alternation.
// No other glob features are supported.
func GlobToRegexp(glob string) (string, error) {
	glob = filepath.ToSlash(glob)

	var b strings.Builder
	b.WriteString("^")

	for i := 0; i < len(glob); i++ {
		ch := glob[i]

		switch ch {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}

		case '{':
			end := strings.IndexByte(glob[i:], '}')
			if end == -1 {
				return "", fmt.Errorf("unterminated brace in glob: %q", glob)
			}
			end += i

			content := glob[i+1 : end]
			if strings.Contains(content, "{") {
				return "", fmt.Errorf("nested braces not supported in glob: %q", glob)
			}

			parts := strings.Split(content, ",")
			if len(parts) == 0 {
				return "", fmt.Errorf("empty alternation in glob: %q", glob)
			}

			b.WriteString("(?:")
			for idx, part := range parts {
				if part == "" {
					return "", fmt.Errorf("empty option in alternation for glob: %q", glob)
				}
				if idx > 0 {
					b.WriteByte('|')
				}
				b.WriteString(regexp.QuoteMeta(part))
			}
			b.WriteString(")")
			i = end

		case '?':
			return "", fmt.Errorf("unsupported glob token '?' in %q", glob)

		default:
			if strings.ContainsRune(`.+()|^$[]{}\`, rune(ch)) {
				b.WriteByte('\\')
			}
			b.WriteByte(ch)
		}
	}

	b.WriteString("$")
	return b.String(), nil
}
