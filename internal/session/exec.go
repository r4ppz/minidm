package session

import "strings"

// ParseExec parses a .desktop Exec value into an argv list following the
// freedesktop.org Desktop Entry Specification.
//
// Whitespace separates arguments unless quoted. Single quotes preserve their
// content literally. Double quotes preserve their content except that a
// backslash may escape the special characters (" $ ` \). A backslash outside
// quotes escapes the next character. Field codes such as %f or %U are
// removed, since a session launcher provides no files or URIs; a literal
// percent sign is written as %%.
func ParseExec(s string) []string {
	var args []string
	var cur strings.Builder
	var inSingle, inDouble bool
	hasArg := false

	flush := func() {
		if hasArg || cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
			hasArg = false
		}
	}

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
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
				if i+1 < len(runes) {
					next := runes[i+1]
					switch next {
					case '"', '$', '`', '\\':
						cur.WriteRune(next)
						i++
					default:
						cur.WriteRune(r)
					}
				} else {
					cur.WriteRune(r)
				}
			case '%':
				if i+1 < len(runes) {
					next := runes[i+1]
					if next == '%' {
						cur.WriteRune('%')
						i++
					} else {
						i++
					}
				}
			default:
				cur.WriteRune(r)
			}

		default:
			switch r {
			case ' ', '\t', '\n':
				flush()
			case '\'':
				inSingle = true
				hasArg = true
			case '"':
				inDouble = true
				hasArg = true
			case '\\':
				if i+1 < len(runes) {
					cur.WriteRune(runes[i+1])
					i++
				} else {
					cur.WriteRune(r)
				}
			case '%':
				if i+1 < len(runes) {
					next := runes[i+1]
					if next == '%' {
						cur.WriteRune('%')
						i++
					} else {
						i++
					}
				}
			default:
				cur.WriteRune(r)
			}
		}
	}
	flush()
	return args
}
