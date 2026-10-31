// Command contrast checks WCAG 2.1 contrast ratios of the colour tokens in
// static/css/00-tokens.css for every theme. Exits non-zero on any failure.
//
//	go run ./tools/contrast [-file static/css/00-tokens.css]
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type pair struct {
	fg, bg string
	min    float64
	note   string
}

// Body-text pairs must reach 4.5:1. The large-text rows document the
// decorative --accent (hero "Muchina" at --fs-3xl), which needs 3:1, and
// the focus-ring rows the 3:1 non-text minimum.
var pairs = []pair{
	{"ink", "bg", 4.5, ""},
	{"ink", "surface", 4.5, ""},
	{"ink", "surface-2", 4.5, ""},
	{"muted", "bg", 4.5, ""},
	{"muted", "surface", 4.5, ""},
	{"muted", "surface-2", 4.5, ""},
	{"accent-text", "bg", 4.5, ""},
	{"accent-text", "surface", 4.5, ""},
	{"accent-text", "surface-2", 4.5, ""},
	{"accent-text", "accent-soft", 4.5, ""},
	{"ink", "accent-soft", 4.5, ""},
	{"on-accent", "accent-strong", 4.5, ""},
	{"on-accent", "accent-strong-hover", 4.5, ""},
	{"accent", "bg", 3.0, "large text"},
	{"accent", "surface-2", 3.0, "large text"},
	// Focus ring (non-text, WCAG 1.4.11: 3:1) on both band colours and on
	// cards/panels.
	{"focus", "bg", 3.0, "focus ring"},
	{"focus", "surface-2", 3.0, "focus ring"},
	{"focus", "surface", 3.0, "focus ring"},
	// Project phones: the placeholder UI's text sits on --screen.
	{"screen-ink", "screen", 4.5, ""},
	{"leaf", "screen", 4.5, ""},
}

// Employment-type colours (the experience section lives on --surface-2):
// each must read on the band and on its own -soft tile.
var typeNames = []string{"internship", "volunteer", "contract", "freelance", "full-time", "part-time", "milestone"}

func init() {
	for _, t := range typeNames {
		pairs = append(pairs,
			pair{"type-" + t, "surface-2", 4.5, ""},
			pair{"type-" + t, "type-" + t + "-soft", 4.5, ""},
		)
	}
}

var themes = []string{"dark", "light"}

var (
	commentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	blockRe   = regexp.MustCompile(`([^{}]+)\{([^{}]*)\}`)
	tokenRe   = regexp.MustCompile(`--([\w-]+)\s*:\s*(#[0-9a-fA-F]{3}(?:[0-9a-fA-F]{3})?)\s*;`)
	refRe     = regexp.MustCompile(`--([\w-]+)\s*:\s*var\(--([\w-]+)\)\s*;`)
)

func main() {
	file := flag.String("file", "static/css/00-tokens.css", "tokens stylesheet")
	flag.Parse()

	src, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	css := commentRe.ReplaceAllString(string(src), "")
	blocks := blockRe.FindAllStringSubmatch(css, -1)

	failed := 0
	fmt.Printf("%-6s  %-40s  %7s  %5s  %s\n", "THEME", "PAIR", "RATIO", "MIN", "RESULT")
	fmt.Println(strings.Repeat("-", 72))
	for _, theme := range themes {
		tokens := resolve(blocks, theme)
		for _, p := range pairs {
			label := p.fg + " / " + p.bg
			if p.note != "" {
				label += " (" + p.note + ")"
			}
			fg, ok1 := tokens[p.fg]
			bg, ok2 := tokens[p.bg]
			if !ok1 || !ok2 {
				fmt.Printf("%-6s  %-40s  %7s  %5.1f  FAIL (token missing)\n", theme, label, "-", p.min)
				failed++
				continue
			}
			r, err := ratio(fg, bg)
			if err != nil {
				fmt.Printf("%-6s  %-40s  %7s  %5.1f  FAIL (%v)\n", theme, label, "-", p.min, err)
				failed++
				continue
			}
			result := "PASS"
			// Compare at the displayed precision so 4.495 never "passes".
			if math.Floor(r*100)/100 < p.min {
				result = "FAIL"
				failed++
			}
			fmt.Printf("%-6s  %-40s  %6.2f:1  %5.1f  %s   %s on %s\n", theme, label, r, p.min, result, fg, bg)
		}
		fmt.Println(strings.Repeat("-", 72))
	}
	if failed > 0 {
		fmt.Printf("%d pair(s) failed\n", failed)
		os.Exit(1)
	}
	fmt.Println("all pairs pass")
}

// resolve mimics the cascade for <html data-theme="theme">: every block whose
// selector list contains :root or [data-theme="theme"] applies, in order.
func resolve(blocks [][]string, theme string) map[string]string {
	want := `[data-theme="` + theme + `"]`
	tokens := map[string]string{}
	refs := map[string]string{}
	for _, b := range blocks {
		applies := false
		for _, sel := range strings.Split(b[1], ",") {
			sel = strings.TrimSpace(sel)
			if sel == ":root" || sel == want {
				applies = true
			}
		}
		if !applies {
			continue
		}
		for _, m := range tokenRe.FindAllStringSubmatch(b[2], -1) {
			tokens[m[1]] = strings.ToUpper(m[2])
		}
		for _, m := range refRe.FindAllStringSubmatch(b[2], -1) {
			refs[m[1]] = m[2]
		}
	}
	// Resolve aliases (--type-internship: var(--accent-text)), a few levels deep.
	for range 4 {
		for name, target := range refs {
			if v, ok := tokens[target]; ok {
				tokens[name] = v
			}
		}
	}
	return tokens
}

func ratio(a, b string) (float64, error) {
	la, err := luminance(a)
	if err != nil {
		return 0, err
	}
	lb, err := luminance(b)
	if err != nil {
		return 0, err
	}
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05), nil
}

// luminance returns WCAG 2.1 relative luminance of a #RGB or #RRGGBB colour.
func luminance(hex string) (float64, error) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 {
		return 0, fmt.Errorf("bad colour %q", hex)
	}
	var ch [3]float64
	for i := range ch {
		v, err := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
		if err != nil {
			return 0, fmt.Errorf("bad colour %q", hex)
		}
		c := float64(v) / 255
		if c <= 0.04045 {
			ch[i] = c / 12.92
		} else {
			ch[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*ch[0] + 0.7152*ch[1] + 0.0722*ch[2], nil
}
