package build

import (
	"bytes"
	"image"
	"image/png"
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
	// No phone number anywhere in the output (Kennedy, 2026-10-30).
	noPhoneNumber(t, out)
	if !strings.Contains(html, `<h2 id="about-title" class="section__title">About me</h2>`) {
		t.Error(`about: the heading should be "About me" (the old tagline is retired)`)
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
	if n := strings.Count(html, `section__action tap-target"`); n != wantActions {
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
	phase5(t, html)
	phase6(t, out, html)

	if len(res.Fingerprinted) != 2 {
		t.Errorf("want 2 fingerprinted files, got %v", res.Fingerprinted)
	}
}

// noPhoneNumber fails if any file in out contains a tel: link or a piece of
// the real phone number.
func noPhoneNumber(t *testing.T, out string) {
	t.Helper()
	err := filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasSuffix(path, ".woff2") || strings.HasSuffix(path, ".jpg") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// html/template writes "+" as "&#43;": check both spellings.
		for _, bad := range []string{"tel:", "703 537", "703537", "+254", "&#43;254"} {
			if strings.Contains(string(data), bad) {
				t.Errorf("%s contains %q: no phone number on the site", path, bad)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// phase5 checks the about, skills, projects, education, contact and footer
// markup against the content.
func phase5(t *testing.T, html string) {
	t.Helper()
	section := func(id string) string {
		i := strings.Index(html, `<section id="`+id+`"`)
		if i < 0 {
			t.Fatalf("no section %s", id)
		}
		j := strings.Index(html[i:], "</section>")
		return html[i : i+j]
	}

	// About: the intro paragraph opens the body, with one <strong> phrase;
	// the derived "Currently" fact comes first.
	about := section("about")
	if !regexp.MustCompile(`<div class="section__body">\s*<p class="intro"`).MatchString(about) {
		t.Error("about: the intro paragraph must be the first thing in the body")
	}
	if !strings.Contains(about, "<strong>reliable, scalable and user-centred applications</strong>") {
		t.Error("about: the bold phrase is missing")
	}
	if strings.Contains(about, `class="eyebrow"`) {
		t.Error("about: no eyebrow expected")
	}
	if !regexp.MustCompile(`<dl class="facts"[^>]*>\s*<div class="facts__item"><dt class="facts__label">Currently</dt><dd class="facts__value">ICT &amp; Digital Innovation<span class="facts__detail">Kenya Flower Council</span>`).MatchString(about) {
		t.Error("about: the first fact should be Currently / the current role")
	}

	// Skills: every pill has an icon; 8 have evidence (buttons + panels).
	skills := section("skills")
	pills := regexp.MustCompile(`class="skill(?: skill--plain)?"[^>]*>(<svg[^>]*class="icon")?`).FindAllStringSubmatch(skills, -1)
	if len(pills) != 29 {
		t.Errorf("skills: want 29 pills, got %d", len(pills))
	}
	for _, m := range pills {
		if m[1] == "" {
			t.Errorf("skills: a pill without an icon: %s", m[0])
		}
	}
	btns := regexp.MustCompile(`<button type="button" class="skill" aria-expanded="true" aria-controls="([^"]+)">`).FindAllStringSubmatch(skills, -1)
	if len(btns) != 8 {
		t.Errorf("skills: want 8 evidence buttons, got %d", len(btns))
	}
	for _, m := range btns {
		if !strings.Contains(skills, `<div class="used" id="`+m[1]+`"`) {
			t.Errorf("skills: no panel %s", m[1])
		}
	}
	for _, href := range regexp.MustCompile(`class="used__link" href="#([^"]+)"`).FindAllStringSubmatch(skills, -1) {
		if !strings.Contains(html, `id="`+href[1]+`"`) {
			t.Errorf("skills: evidence link #%s has no target", href[1])
		}
	}
	if strings.Contains(skills, " hidden") {
		t.Error("skills: panels must be visible without JS (main.js hides them)")
	}

	// Projects: placeholders render (no screenshots yet), the stage is hidden
	// from assistive tech, and the linked card is clickable as a whole.
	projects := section("projects")
	if strings.Count(projects, `<div class="project__stage" aria-hidden="true">`) != 2 {
		t.Error("projects: want 2 aria-hidden placeholder stages")
	}
	if !strings.Contains(projects, `class="project card lift card--interactive project--staged"`) ||
		!strings.Contains(projects, `class="project__link card__link"`) {
		t.Error("projects: the linked card should be interactive with a stretched link")
	}
	if !strings.Contains(projects, `<strong>100&#43;</strong> active users`) {
		t.Error("projects: metric missing")
	}

	// Education: degree card, grades, meters from LinkedIn levels.
	edu := section("education")
	for _, want := range []string{
		`<p class="honours">First Class Honours<small>Graduated magna cum laude, Class of 2025</small></p>`,
		`<time datetime="2025-12">Dec 2025</time>`,
		`<div><dt>Mean grade</dt><dd>A−</dd></div>`,
		`role="img" aria-label="Level 3 of 5"><i class="on"></i><i class="on"></i><i class="on"></i><i></i><i></i></span>`,
		`aria-label="Level 5 of 5"`,
	} {
		if !strings.Contains(edu, want) {
			t.Errorf("education: missing %s", want)
		}
	}

	// Contact: channels in order, the clock with its zone, the copy button.
	contact := section("contact")
	if got := regexp.MustCompile(`class="channels__name">([^<]+)<`).FindAllStringSubmatch(contact, -1); len(got) != 3 || got[0][1] != "LinkedIn" || got[2][1] != "Google Play" {
		t.Errorf("contact: channels = %v", got)
	}
	if !regexp.MustCompile(`<time datetime="[0-9-]{10}T[0-9:]{5}&#43;03:00" data-tz="Africa/Nairobi">[0-2][0-9]:[0-5][0-9]</time> here \(EAT, UTC&#43;3\)`).MatchString(contact) {
		t.Error("contact: build-time Nairobi clock missing")
	}
	if !strings.Contains(contact, `data-copy="muchina537@gmail.com"`) {
		t.Error("contact: copy button missing")
	}

	// Experience: job anchors for the evidence links.
	for _, id := range []string{"job-kenya-flower-council", "job-kenya-roads-board", "job-ieee-tech4good", "job-ieee-moi-university-student-branch"} {
		if !strings.Contains(html, `<article class="job" id="`+id+`">`) {
			t.Errorf("experience: no job article %s", id)
		}
	}

	// Footer: the credit line (no source link) and Back to top.
	if !regexp.MustCompile(`<p class="site-footer__built">© <span data-year>[0-9]{4}</span>. Built with Golang</p>`).MatchString(html) {
		t.Error("footer: credit line should be just © year. Built with Golang")
	}
	if !strings.Contains(html, `<a class="to-top" href="#top">`) {
		t.Error("footer: Back to top missing")
	}
}

// A screenshot under static/ replaces the placeholder and is not hidden.
func TestScreenshotReplacesPlaceholder(t *testing.T) {
	tmp := t.TempDir()
	static := filepath.Join(tmp, "static")
	if err := os.CopyFS(static, os.DirFS("../../static")); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 108, 222))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(static, "img", "projects", "apple-care.png"), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(tmp, "dist")
	res, err := Run(Options{Content: "../../content/site.yaml", Templates: "../../templates", Static: static, Out: out})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	if !regexp.MustCompile(`<div class="project__stage">\s*<div class="phone"><div class="phone__screen"><span class="phone__notch"></span>\s*<img class="phone__image" src="/img/projects/apple-care.png" width="108" height="222" alt="Screenshot of Apple Care" loading="lazy" decoding="async">`).MatchString(html) {
		t.Error("the screenshot should replace the scanner placeholder, with its size and alt")
	}
	if strings.Contains(html, `class="ui__finder"`) {
		t.Error("the scanner placeholder should be gone")
	}
	if strings.Count(html, `<div class="project__stage" aria-hidden="true">`) != 1 {
		t.Error("only the journal placeholder stage stays aria-hidden")
	}
	if len(res.Warnings) != 1 {
		t.Errorf("want 1 screenshot warning (the journal), got %v", res.Warnings)
	}
}

// phase6 checks the navigation, deep-link and accessibility markup.
func phase6(t *testing.T, out, html string) {
	t.Helper()

	// One polite live region per page (copy email, legend, theme), carrying
	// the theme wording; <main> can take focus from the skip link.
	if n := strings.Count(html, `role="status"`); n != 1 {
		t.Errorf("want exactly 1 live region, got %d", n)
	}
	if !regexp.MustCompile(`<div id="sr-status" class="sr-only" role="status" data-status data-dark="Dark theme" data-light="Light theme" data-system="System theme, currently \{theme\}"`).MatchString(html) {
		t.Error("#sr-status live region (with the theme announcements) missing")
	}
	if !strings.Contains(html, `<main id="main" tabindex="-1">`) {
		t.Error(`<main id="main" tabindex="-1"> missing`)
	}

	// The scrollspy indicator sits in the desktop nav.
	if n := strings.Count(html, `class="nav__indicator" aria-hidden="true" data-nav-indicator`); n != 1 {
		t.Errorf("want 1 nav indicator, got %d", n)
	}

	// No scroll reveals left anywhere: not in the markup, CSS or JS.
	err := filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !(strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js")) {
			return err
		}
		data, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(data), "data-reveal") {
			t.Errorf("%s still mentions data-reveal", path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	// Skill buttons: the count is part of the name (hidden text), the badge is aria-hidden.
	for _, want := range []string{
		`Kotlin<span class="sr-only">, used in 1 place</span> <span class="skill__count" aria-hidden="true">1</span>`,
		`Business Central<span class="sr-only">, used at 1 place</span>`,
		`Team leadership<span class="sr-only">, shown at 2 places</span> <span class="skill__count" aria-hidden="true">2</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("skill button name missing: %s", want)
		}
	}
	if got := strings.Count(html, `<span class="skill__count" aria-hidden="true">`); got != 8 {
		t.Errorf("want 8 aria-hidden skill counts, got %d", got)
	}

	// Legend buttons say what they show.
	says := regexp.MustCompile(`class="legend__item"[^>]*data-announce="([^"]+)"`).FindAllStringSubmatch(html, -1)
	var got []string
	for _, m := range says {
		got = append(got, m[1])
	}
	if strings.Join(got, "; ") != "Showing all roles; Showing internship roles; Showing volunteer roles; Showing milestones" {
		t.Errorf("legend announcements = %v", got)
	}

	// Every link that leaves the site has rel="noopener".
	for _, m := range regexp.MustCompile(`<a [^>]*href="https://[^"]*"[^>]*>`).FindAllString(html, -1) {
		if !regexp.MustCompile(`rel="[^"]*\bnoopener\b`).MatchString(m) {
			t.Errorf("external link without noopener: %s", m)
		}
	}

	// Links to the résumé PDF say so to screen readers.
	if n := strings.Count(html, `<span class="sr-only"> (PDF)</span>`); n != 2 {
		t.Errorf(`want 2 hidden "(PDF)" suffixes (hero, experience), got %d`, n)
	}
	if !regexp.MustCompile(`<a class="btn btn--ghost press" href="/Kennedy_Muchina_Resume\.pdf" download>`).MatchString(html) {
		t.Error("the résumé button lost its download attribute")
	}
}
