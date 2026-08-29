package session

import "strings"

// ParseExec parses a .desktop Exec value into an argv list per the
// freedesktop.org Desktop Entry Specification. Field codes (%f, %U, etc.)
// are dropped since a session launcher provides no files or URIs.
func ParseExec(s string) []string {
	var args []string
	var cur strings.Builder
	var inSingle, inDouble bool
	hasToken := false

	flush := func() {
		if hasToken || cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
			hasToken = false
		}
	}

	runes := []rune(s)
	i := 0
	for i < len(runes) {
		r := runes[i]

		switch {
		case inSingle:
			if r == '\'' {
				inSingle = false
			} else {
				cur.WriteRune(r)
			}

		case inDouble:
			switch r {
			case '"':
				inDouble = false
			case '\\':
				i = escapeInDouble(runes, i, &cur)
			case '%':
				i = handleFieldCode(runes, i, &cur)
			default:
				cur.WriteRune(r)
			}

		default:
			switch r {
			case ' ', '\t', '\n':
				flush()
			case '\'':
				inSingle = true
				hasToken = true
			case '"':
				inDouble = true
				hasToken = true
			case '\\':
				i = escapeRaw(runes, i, &cur)
			case '%':
				i = handleFieldCode(runes, i, &cur)
			default:
				cur.WriteRune(r)
			}
		}
		i++
	}
	flush()
	return args
}

// Only " $ ` \ are unescaped inside double quotes; any other char after
// backslash is kept literally.
func escapeInDouble(runes []rune, i int, cur *strings.Builder) int {
	if i+1 >= len(runes) {
		cur.WriteRune(runes[i])
		return i
	}
	next := runes[i+1]
	switch next {
	case '"', '$', '`', '\\':
		cur.WriteRune(next)
	default:
		cur.WriteRune(runes[i])
	}
	return i + 1
}

func escapeRaw(runes []rune, i int, cur *strings.Builder) int {
	if i+1 >= len(runes) {
		cur.WriteRune(runes[i])
		return i
	}
	cur.WriteRune(runes[i+1])
	return i + 1
}

// %% produces a literal percent; any other field code is dropped.
func handleFieldCode(runes []rune, i int, cur *strings.Builder) int {
	if i+1 >= len(runes) {
		return i
	}
	if runes[i+1] == '%' {
		cur.WriteRune('%')
	}
	return i + 1
}
