package build

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/kmuchina/portfolio/internal/site"
)

func TestFullBuild(t *testing.T) {
	out := filepath.Join(t.TempDir(), "dist")
	res, err := Run(Options{
		Content:   "../../content/site.yaml",
		Templates: "../../templates",
		Static:    "../../static",
		Out:       out,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)

	if n := strings.Count(html, "<h1"); n != 1 {
		t.Errorf("want exactly one <h1, got %d", n)
	}
	for _, id := range site.SectionIDs {
		if !strings.Contains(html, `<section id="`+id+`"`) {
			t.Errorf("missing section id %q", id)
		}
	}
	if !strings.Contains(html, "muchina537@gmail.com") {
		t.Error("contact email missing")
	}
	if strings.Contains(html, "kmuchina22") {
		t.Error("retired email present")
	}
	// show_phone is off in the real content: no phone link may render.
	if strings.Contains(html, "tel:") {
		t.Error("a tel: link is present although site.show_phone is false")
	}

	// Fingerprinted assets are referenced and exist.
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`href="(/css/site\.[0-9a-f]{8}\.css)"`),
		regexp.MustCompile(`src="(/js/main\.[0-9a-f]{8}\.js)"`),
	} {
		m := re.FindStringSubmatch(html)
		if m == nil {
			t.Errorf("no reference matching %s", re)
			continue
		}
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(m[1]))); err != nil {
			t.Errorf("%s referenced but not written: %v", m[1], err)
		}
	}

	// Every header and menu nav link points at an element that exists.
	navRe := regexp.MustCompile(`class="(?:nav__link[^"]*|menu__link)" href="#([^"]+)"`)
	navLinks := navRe.FindAllStringSubmatch(html, -1)
	if want := 2 * 6; len(navLinks) != want { // About…Contact, header + menu
		t.Errorf("want %d nav links (header + menu), got %d", want, len(navLinks))
	}
	for _, m := range navLinks {
		if !strings.Contains(html, `id="`+m[1]+`"`) {
			t.Errorf("nav link #%s has no matching id", m[1])
		}
	}
	if strings.Contains(html, `href="#top">Home<`) {
		t.Error(`"Home" should not be in the nav (the brand links to #top)`)
	}
	if n := strings.Count(html, `data-theme-toggle`); n != 2 {
		t.Errorf("want 2 theme toggles (header + menu), got %d", n)
	}

	// Section actions render as arrow links in the spine asides; the résumé
	// is optional (empty resume_pdf hides it). Both follow the content.
	s, err := site.Load("../../content/site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	wantActions := 0
	for _, a := range []*site.Action{s.Experience.Action, s.Projects.Action} {
		if a != nil {
			wantActions++
		}
	}
	if n := strings.Count(html, `class="link-arrow section__action tap-target"`); n != wantActions {
		t.Errorf("want %d section action links, got %d", wantActions, n)
	}
	if hasResume := strings.Contains(html, `download>`); hasResume != (s.Person.ResumePDF != "") {
		t.Errorf("résumé download button rendered = %v, but resume_pdf = %q", hasResume, s.Person.ResumePDF)
	}

	// The career graph: 5 rows, newest first, with the right kinds.
	rowRe := regexp.MustCompile(`<li class="career__row career__row--([a-z-]+)">[\s\S]*?<h3 class="job__title">([^<]+)</h3>`)
	var rows []string
	for _, m := range rowRe.FindAllStringSubmatch(html, -1) {
		rows = append(rows, m[1]+":"+m[2])
	}
	wantRows := "internship:ICT &amp; Digital Innovation | milestone:Graduated, BSc Computer Science | " +
		"internship:ICT Officer | volunteer:Project Lead &amp; Tech Lead | volunteer:Chairperson"
	if got := strings.Join(rows, " | "); got != wantRows {
		t.Errorf("career rows:\n got  %s\n want %s", got, wantRows)
	}
	if !strings.Contains(html, `<span class="summary__text">roles since 2023, up to three at once</span>`) {
		t.Error("career summary text missing or wrong")
	}
	if n := strings.Count(html, `<svg class="curve"`); n != 1 {
		t.Errorf("want 1 merge curve, got %d", n)
	}
	if !strings.Contains(html, `achieving <strong>99% network uptime</strong> across`) {
		t.Error("**bold** in bullets not rendered as <strong>")
	}
	if regexp.MustCompile(`(?i)class="job__title">[^<]*\bintern`).MatchString(html) {
		t.Error(`a job title contains "intern"`)
	}

	// The cascade-layer order is declared before any layer block.
	if m := regexp.MustCompile(`href="(/css/site\.[0-9a-f]{8}\.css)"`).FindStringSubmatch(html); m != nil {
		css, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(m[1])))
		if err != nil {
			t.Fatal(err)
		}
		decl := strings.Index(string(css), "@layer reset, tokens, base, layout, components, sections, utilities, motion;")
		first := strings.Index(string(css), "@layer tokens {")
		if decl < 0 || first < 0 || decl > first {
			t.Error("@layer order declaration must come before the first layer block")
		}
	}

	// theme-init.js is inlined in <head>, not shipped as a file.
	head := html[:strings.Index(html, "</head>")]
	if !strings.Contains(head, `localStorage.getItem("theme")`) {
		t.Error("theme-init script not inlined in <head>")
	}
	if _, err := os.Stat(filepath.Join(out, "js", "theme-init.js")); err == nil {
		t.Error("theme-init.js should not be copied to the output")
	}

	if _, err := os.Stat(filepath.Join(out, "design")); err == nil {
		t.Error("design/ must not be in the output")
	}
	for _, f := range []string{"404.html", "styleguide.html", "favicon.svg"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}

	// GitHub Pages: CNAME is the host of site.url; .nojekyll exists and is empty.
	if cname, err := os.ReadFile(filepath.Join(out, "CNAME")); err != nil {
		t.Errorf("missing CNAME: %v", err)
	} else if got := strings.TrimSpace(string(cname)); got != "kmuchina.com" {
		t.Errorf("CNAME = %q, want %q", got, "kmuchina.com")
	}
	if fi, err := os.Stat(filepath.Join(out, ".nojekyll")); err != nil {
		t.Errorf("missing .nojekyll: %v", err)
	} else if fi.Size() != 0 {
		t.Errorf(".nojekyll should be empty, got %d bytes", fi.Size())
	}
	if len(res.Fingerprinted) != 2 {
		t.Errorf("want 2 fingerprinted files, got %v", res.Fingerprinted)
	}
}
