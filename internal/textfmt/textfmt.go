// Package textfmt formats the tiny inline syntax allowed in content: **bold**.
// No Markdown dependency: everything is HTML-escaped first, then each closed
// **…** pair becomes <strong>…</strong>. An unclosed ** stays as literal text.
package textfmt

import (
	"html"
	"strings"
)

// Bold returns s as safe HTML with **x** turned into <strong>x</strong>.
func Bold(s string) string {
	parts := strings.Split(s, "**")
	var b strings.Builder
	for i, p := range parts {
		esc := html.EscapeString(p)
		switch {
		case i%2 == 0:
			b.WriteString(esc) // outside bold
		case i == len(parts)-1:
			b.WriteString("**" + esc) // unclosed: the ** was literal
		default:
			b.WriteString("<strong>" + esc + "</strong>")
		}
	}
	return b.String()
}
