package assets

// MinifyCSS is a deliberately small, safe CSS minifier:
//   - comments become a single space (so `0/**/auto` never fuses to `0auto`);
//   - runs of whitespace collapse to one space;
//   - whitespace next to { } ; , and after : is dropped, and so is a ;
//     just before }.
//
// It never touches the inside of a string, and never removes the space
// before a ":" or around combinators and operators: `.a :is(b)` (a
// descendant) and `.a:is(b)` differ, and calc() needs the spaces around + and -.
func MinifyCSS(src []byte) []byte {
	out := make([]byte, 0, len(src))
	pendingSpace := false
	drops := func(c byte) bool { return c == '{' || c == '}' || c == ';' || c == ',' }
	emit := func(c byte) {
		if pendingSpace {
			if len(out) > 0 && !drops(out[len(out)-1]) && out[len(out)-1] != ':' && !drops(c) {
				out = append(out, ' ')
			}
			pendingSpace = false
		}
		if c == '}' && len(out) > 0 && out[len(out)-1] == ';' {
			out = out[:len(out)-1]
		}
		out = append(out, c)
	}
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch {
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			end := indexFrom(src, i+2, "*/")
			if end < 0 {
				i = len(src)
			} else {
				i = end + 1
			}
			pendingSpace = true
		case c == '"' || c == '\'':
			emit(c)
			for i++; i < len(src); i++ {
				out = append(out, src[i])
				if src[i] == '\\' && i+1 < len(src) {
					i++
					out = append(out, src[i])
					continue
				}
				if src[i] == c {
					break
				}
			}
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f':
			pendingSpace = true
		default:
			emit(c)
		}
	}
	return out
}

func indexFrom(b []byte, from int, sub string) int {
	for i := from; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == sub {
			return i
		}
	}
	return -1
}
