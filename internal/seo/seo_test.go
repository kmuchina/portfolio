package seo

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kmuchina/portfolio/internal/site"
)

func load(t *testing.T) *site.Site {
	t.Helper()
	s, err := site.Load("../../content/site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestPersonFromContent(t *testing.T) {
	s := load(t)
	p := NewPerson(s, "https://kmuchina.com/img/profile-720.abc.jpg")
	if p.WorksFor == nil || p.WorksFor.Name != "Kenya Flower Council" {
		t.Errorf("worksFor = %+v", p.WorksFor)
	}
	if p.AlumniOf == nil || p.AlumniOf.Type != "CollegeOrUniversity" || p.AlumniOf.Name != "Moi University" {
		t.Errorf("alumniOf = %+v", p.AlumniOf)
	}
	if got := strings.Join(p.KnowsAbout, ", "); got != "Kotlin, Jetpack Compose, JavaScript, Node.js, Express, HTML5, CSS3, C++, Python, TensorFlow Lite, CNN" {
		t.Errorf("knowsAbout = %s", got)
	}
	if len(p.SameAs) != 3 || p.Email != "mailto:muchina537@gmail.com" || p.URL != "https://kmuchina.com/" {
		t.Errorf("person = %+v", p)
	}
}

// No current role → no worksFor; the org is cut at its first comma.
func TestPersonWorksFor(t *testing.T) {
	s := load(t)
	for i := range s.Experience.Items {
		s.Experience.Items[i].End = "2025-01"
	}
	if p := NewPerson(s, ""); p.WorksFor != nil {
		t.Errorf("no current role: worksFor should be absent, got %+v", p.WorksFor)
	}
	s.Experience.Items[1].End = site.Present // "Kenya Roads Board, Ministry of Roads & Transport"
	if p := NewPerson(s, ""); p.WorksFor == nil || p.WorksFor.Name != "Kenya Roads Board" {
		t.Errorf("worksFor = %+v", p.WorksFor)
	}
}

// The JSON can't close its <script> element, and it parses back.
func TestJSONLDEscapes(t *testing.T) {
	s := load(t)
	s.Person.Role = `Engineer </script><script>alert(1)</script>`
	b, err := JSONLD(NewPerson(s, ""))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "</script>") || strings.Contains(string(b), "<") {
		t.Errorf("JSON-LD must escape <: %s", b)
	}
	var back map[string]any
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back["jobTitle"] != s.Person.Role {
		t.Errorf("round trip: %v", back["jobTitle"])
	}
	if _, ok := back["telephone"]; ok {
		t.Error("no telephone, ever")
	}
}

func TestCrawlFiles(t *testing.T) {
	if got := string(Robots("https://kmuchina.com")); got != "User-agent: *\nAllow: /\nDisallow: /styleguide.html\n\nSitemap: https://kmuchina.com/sitemap.xml\n" {
		t.Errorf("robots.txt = %q", got)
	}
	b, err := Sitemap("https://kmuchina.com/", time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	want := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://kmuchina.com/</loc>
    <lastmod>2026-10-31</lastmod>
  </url>
</urlset>
`
	if string(b) != want {
		t.Errorf("sitemap.xml =\n%s", b)
	}
	if HomeURL("https://kmuchina.com") != "https://kmuchina.com/" || Abs("https://kmuchina.com/", "/og.png") != "https://kmuchina.com/og.png" {
		t.Error("HomeURL / Abs")
	}
}
