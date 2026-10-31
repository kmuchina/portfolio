package assets

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestMinifyCSS(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"a {\n  color: red;\n}\n", "a{color:red}"},
		{"/* c */ a , b { margin: 0 auto ; }", "a,b{margin:0 auto}"},
		{"a{margin:0/**/auto}", "a{margin:0 auto}"},                                              // a comment never fuses tokens
		{".a :is(.b) { x: y }", ".a :is(.b){x:y}"},                                                // descendant + pseudo keeps its space
		{"a { width: calc(100% - var(--x) + 2px); }", "a{width:calc(100% - var(--x) + 2px)}"},     // calc operators
		{`a::after { content: " (" attr(href) ")"; }`, `a::after{content:" (" attr(href) ")"}`},   // strings untouched
		{`a { content: "a  /* not a comment */  b"; }`, `a{content:"a  /* not a comment */  b"}`}, // comment markers in strings
		{`a { content: 'it\'s ; }'; }`, `a{content:'it\'s ; }'}`},                                // escaped quote
		{"@media (min-width: 900px) {\n  a { b: c }\n}", "@media (min-width:900px){a{b:c}}"},
		{"@layer reset, tokens;\n", "@layer reset,tokens;"},
		{"a { --e: ; }", "a{--e:}"},
	} {
		if got := string(MinifyCSS([]byte(c.in))); got != c.want {
			t.Errorf("MinifyCSS(%q)\n got %q\nwant %q", c.in, got, c.want)
		}
	}
}

// The real stylesheet keeps every string, every rule and its layer order.
func TestMinifyRealCSS(t *testing.T) {
	files, _ := filepath.Glob("../../static/css/*.css")
	sort.Strings(files)
	var all strings.Builder
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		all.Write(b)
		all.WriteString("\n")
	}
	src := all.String()
	min := string(MinifyCSS([]byte(src)))
	if len(min) > len(src)*7/10 {
		t.Errorf("expected at least 30%% smaller: %d → %d bytes", len(src), len(min))
	}
	if strings.Count(min, "{") != strings.Count(min, "}") {
		t.Error("unbalanced braces after minifying")
	}
	// Strings and declaration counts survive (comments stripped first).
	noComments := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(src, "")
	strRe := regexp.MustCompile(`"[^"\n]*"`)
	if a, b := strRe.FindAllString(noComments, -1), strRe.FindAllString(min, -1); strings.Join(a, "|") != strings.Join(b, "|") {
		t.Errorf("strings changed: %d before, %d after", len(a), len(b))
	}
	if !strings.HasPrefix(min, "@layer reset,tokens,base,layout,components,sections,utilities,") {
		t.Errorf("layer order should open the file: %.80s", min)
	}
}

// The print palette in 00-tokens.css is a copy of the light theme (CSS
// can't apply a selector only in print); it must never drift.
func TestPrintTokensMatchLight(t *testing.T) {
	data, err := os.ReadFile("../../static/css/00-tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	css := tokenCommentRe.ReplaceAllString(string(data), "")
	decls := func(selector string) string {
		for _, b := range tokenBlockRe.FindAllStringSubmatch(css, -1) {
			if strings.TrimSpace(b[1]) == selector {
				return strings.Join(strings.Fields(b[2]), " ")
			}
		}
		t.Fatalf("no %s block in 00-tokens.css", selector)
		return ""
	}
	light, print := decls(`[data-theme="light"]`), decls(`:root[data-theme]`)
	if light != print {
		t.Errorf("the print block differs from the light theme:\nlight: %s\nprint: %s", light, print)
	}
}
