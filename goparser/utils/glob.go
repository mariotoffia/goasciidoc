package utils

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

// GlobToRegexp converts a shell-style glob pattern to a compiled *regexp.Regexp.
//
// The returned regexp is anchored: it matches the entire input string (not a substring).
// The pattern is interpreted in a filesystem/path context, similar to common glob
// implementations (Bash, Python's glob, Go's filepath.Match) with some extensions.
//
// Supported glob features:
//
//   - Literals:  ordinary characters match themselves.
//   - *         matches zero or more non-separator characters.
//   - ?         matches exactly one non-separator character.
//   - **        matches zero or more characters, including path separators
//     (i.e. across directory boundaries). There is no special requirement
//     that it be a whole path segment.
//   - [abc]     character class: matches one character from the set.
//   - [a-z]     range inside a character class.
//   - [!abc]    negated character class (also [^abc]).
//   - [[:...:]] POSIX character classes inside [...], e.g. [[:digit:]].
//   - {a,b,c}   alternation. This is translated into regex alternation inside a single
//     pattern (e.g. foo.{png,jpg} → ^foo\.(?:png|jpg)$).
//     Nested braces are supported, e.g. {a,{b,c}}.
//     Empty options are allowed: {,bak} means “empty string or bak”.
//
// Escaping:
//
//   - On Unix-like systems (GOOS != "windows"):
//
//   - A backslash '\' escapes the next character outside of [...].
//     For example, \* matches a literal '*'.
//
//   - Inside a character class [...], '\' escapes the next character so it
//     is treated literally in the class (e.g. [\-\]] matches '-' or ']').
//
//   - A trailing '\' at the end of the pattern is an error.
//
//   - On Windows (GOOS == "windows"):
//
//   - '\' is treated as a path separator, not as an escape (matching Go's
//     filepath.Match behavior).
//
//   - To match a literal glob metacharacter you must use character classes
//     (e.g. [*] to match a literal '*').
//
// Path separator semantics:
//
//   - On Unix-like systems, '/' is the only path separator.
//   - '*', '?', and character classes never match '/'.
//   - On Windows, both '/' and '\' are treated as path separators.
//   - '*', '?', and character classes never match '/' or '\'.
//   - '/' and '\' in the pattern both match “a separator” in the regexp.
//
// Anchoring:
//
//   - The generated regexp is wrapped with \A and \z, so it must match the entire
//     string. It does not search for a substring match.
//
// Not supported:
//
//   - Extended globs like @(pattern), !(pattern), +(pattern), *(pattern) (Bash extglob).
//   - Tilde expansion (~user), environment variable expansion, or any filesystem I/O.
//   - Escaping using anything other than '\' on Unix or character classes.
//
// Errors:
//
//   - Unmatched '[' or ']'.
//   - Empty character class ([] or [!]).
//   - Unmatched '{' or '}'.
//   - Trailing '\' on Unix.
//   - Regex compilation errors (should not normally occur).
//
// Example:
//
//	re, err := GlobToRegexp("src/**/test_{small,large}.go")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(re.MatchString("src/test_small.go"))        // true
//	fmt.Println(re.MatchString("src/pkg/test_large.go"))    // true
//	fmt.Println(re.MatchString("src/pkg/other.go"))         // false
func GlobToRegexp(pattern string) (*regexp.Regexp, error) {
	style := detectOSStyle()

	body, err := globToRegexBody(pattern, style)
	if err != nil {
		return nil, err
	}

	// Anchor to entire string using \A and \z to avoid multiline surprises.
	src := `\A` + body + `\z`

	re, err := regexp.Compile(src)
	if err != nil {
		return nil, fmt.Errorf("globutil: compiled regex invalid: %w (src=%q)", err, src)
	}
	return re, nil
}

type osStyle struct {
	windows    bool
	sepPattern string // regex fragment that matches a single path separator
	nonSep     string // regex fragment that matches a single non-separator character
}

// detectOSStyle returns separator semantics based on runtime.GOOS.
//
// On Unix:   sep = '/',   nonSep = [^/]
// On Windows: sep = [/\\], nonSep = [^/\\]
func detectOSStyle() osStyle {
	if runtime.GOOS == "windows" {
		return osStyle{
			windows:    true,
			sepPattern: `[\\/]]`,
			nonSep:     `[^\\/]]`,
		}
	}
	return osStyle{
		windows:    false,
		sepPattern: `/`,
		nonSep:     `[^/]`,
	}
}

// globToRegexBody converts a glob sub-pattern (no anchors) into a regex body string.
// It is recursive (for nested brace alternation).
func globToRegexBody(pat string, style osStyle) (string, error) {
	var b strings.Builder

	for i := 0; i < len(pat); {
		ch := pat[i]

		switch ch {
		case '\\':
			if style.windows {
				// On Windows, backslash is always a separator.
				b.WriteString(style.sepPattern)
				i++
			} else {
				// Escape next character literally.
				if i+1 >= len(pat) {
					return "", fmt.Errorf("globutil: trailing backslash escape in %q", pat)
				}
				escapeRegexLiteral(&b, pat[i+1])
				i += 2
			}

		case '/':
			// Forward slash: path separator on all platforms.
			b.WriteString(style.sepPattern)
			i++

		case '[':
			frag, next, err := parseCharClass(pat, i, style)
			if err != nil {
				return "", err
			}
			b.WriteString(frag)
			i = next

		case '{':
			frag, next, err := parseBraceGroup(pat, i, style)
			if err != nil {
				return "", err
			}
			b.WriteString(frag)
			i = next

		case '*':
			// Check for globstar '**'
			if i+1 < len(pat) && pat[i+1] == '*' {
				// Globstar: match any characters, including separators.
				// (?s:.*) enables dot to match newlines as well.
				b.WriteString(`(?s:.*)`)
				i += 2
			} else {
				// Single '*': zero or more non-separator characters.
				b.WriteString(style.nonSep)
				b.WriteByte('*')
				i++
			}

		case '?':
			// Single '?': exactly one non-separator character.
			b.WriteString(style.nonSep)
			i++

		default:
			escapeRegexLiteral(&b, ch)
			i++
		}
	}

	return b.String(), nil
}

// escapeRegexLiteral writes ch as a literal in a regexp, escaping if needed.
func escapeRegexLiteral(b *strings.Builder, ch byte) {
	switch ch {
	case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']', '?', '*', '\\':
		b.WriteByte('\\')
	}
	b.WriteByte(ch)
}

// parseCharClass parses a character class starting at pat[start] == '['.
//
// It returns the regexp fragment, and the index of the next character after the class.
func parseCharClass(pat string, start int, style osStyle) (string, int, error) {
	n := len(pat)
	j := start + 1
	if j >= n {
		return "", 0, fmt.Errorf("globutil: unterminated character class in %q", pat)
	}

	negated := false
	if pat[j] == '!' || pat[j] == '^' {
		negated = true
		j++
	}
	if j >= n {
		return "", 0, fmt.Errorf("globutil: unterminated character class in %q", pat)
	}
	if pat[j] == ']' {
		// [] or [!] is considered invalid (empty class).
		return "", 0, fmt.Errorf("globutil: empty character class in %q", pat)
	}

	// Find closing ']'
	k := j
	for ; k < n; k++ {
		c := pat[k]
		if !style.windows && c == '\\' && k+1 < n {
			// Skip escaped character inside class on Unix.
			k++
			continue
		}
		if c == ']' {
			break
		}
	}
	if k >= n {
		return "", 0, fmt.Errorf("globutil: missing ']' in character class %q", pat)
	}
	content := pat[j:k]
	next := k + 1

	var b strings.Builder
	b.WriteByte('[')
	if negated {
		b.WriteByte('^')
	}

	// Emit content, handling escapes, ranges, POSIX classes.
	for i := 0; i < len(content); i++ {
		c := content[i]

		if !style.windows && c == '\\' && i+1 < len(content) {
			// Escaped character in class.
			nextChar := content[i+1]
			if nextChar == '\\' || nextChar == ']' || nextChar == '-' || nextChar == '^' {
				b.WriteByte('\\')
			}
			b.WriteByte(nextChar)
			i++
			continue
		}

		// POSIX class, e.g. [:digit:]
		if c == '[' && i+1 < len(content) && content[i+1] == ':' {
			end := strings.Index(content[i:], ":]")
			if end > 0 {
				// Copy [:name:] verbatim.
				b.WriteString(content[i : i+end+2])
				i += end + 1
				continue
			}
			// Fall through: treat '[' literally if malformed.
		}

		if c == '-' {
			// Literal '-' if at start or end; otherwise allow as range operator.
			if i == 0 || i == len(content)-1 {
				b.WriteString(`\-`)
			} else {
				b.WriteByte('-')
			}
		} else {
			// Escape regex specials inside the class.
			if c == ']' || c == '\\' || c == '^' {
				b.WriteByte('\\')
			}
			b.WriteByte(c)
		}
	}

	b.WriteByte(']')
	return b.String(), next, nil
}

// parseBraceGroup parses an alternation group starting at pat[start] == '{'.
//
// It supports nested braces and splits on commas at the top brace depth, ignoring
// commas inside nested {...} or [...].
//
// Returns the regexp fragment "(?:alt1|alt2|...)" and the index of the next
// character after the closing '}'.
func parseBraceGroup(pat string, start int, style osStyle) (string, int, error) {
	n := len(pat)
	depth := 1
	inClass := false
	j := start + 1

	for ; j < n; j++ {
		c := pat[j]

		if !style.windows && c == '\\' && j+1 < n {
			// Skip escaped char.
			j++
			continue
		}

		switch c {
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '{':
			if !inClass {
				depth++
			}
		case '}':
			if !inClass {
				depth--
				if depth == 0 {
					goto foundEnd
				}
			}
		}
	}

	// If we reach here, no matching '}'.
	return "", 0, fmt.Errorf("globutil: unmatched '{' in %q", pat)

foundEnd:
	end := j
	inner := pat[start+1 : end]
	next := end + 1

	// Split inner on commas at depth 0, outside [...].
	var parts []string
	partStart := 0
	depth = 0
	inClass = false

	for i := 0; i < len(inner); i++ {
		c := inner[i]

		if !style.windows && c == '\\' && i+1 < len(inner) {
			// Skip escaped char.
			i++
			continue
		}

		switch c {
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '{':
			if !inClass {
				depth++
			}
		case '}':
			if !inClass {
				depth--
			}
		case ',':
			if !inClass && depth == 0 {
				parts = append(parts, inner[partStart:i])
				partStart = i + 1
			}
		}
	}
	parts = append(parts, inner[partStart:])

	// Convert each alternative recursively.
	if len(parts) == 0 {
		// Should not happen (we always append at least one).
		return "", 0, fmt.Errorf("globutil: empty brace group in %q", pat)
	}

	var b strings.Builder
	b.WriteString("(?:")
	for idx, p := range parts {
		if idx > 0 {
			b.WriteByte('|')
		}
		altRe, err := globToRegexBody(p, style)
		if err != nil {
			return "", 0, err
		}
		b.WriteString(altRe)
	}
	b.WriteByte(')')

	return b.String(), next, nil
}
