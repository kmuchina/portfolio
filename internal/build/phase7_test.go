package build

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// phase7 checks the production polish: responsive images, the head meta,
// the JSON-LD, the crawl files, the icons and manifest, minified CSS and
// the 404 page.
func phase7(t *testing.T, out, html string) {
	t.Helper()
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			t.Errorf("dist/%s: %v", name, err)
		}
		return string(b)
	}

	// The files GitHub Pages, crawlers and browsers ask for by name.
	for _, f := range []string{"robots.txt", "sitemap.xml", "og.png", "favicon.svg", "favicon.ico", "apple-touch-icon.png", "site.webmanifest", "404.html", "CNAME", ".nojekyll"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("dist/%s missing", f)
		}
	}

	// Responsive photo: every width, hashed names, never the original.
	if _, err := os.Stat(filepath.Join(out, "img", "profile.jpg")); err == nil {
		t.Error("the original profile.jpg must not be copied to dist/")
	}
	for _, w := range PhotoWidths {
		if m, _ := filepath.Glob(filepath.Join(out, "img", "profile-"+itoa(w)+".*.jpg")); len(m) != 1 {
			t.Errorf("want one img/profile-%d.<hash>.jpg, got %v", w, m)
		}
	}
	hero := regexp.MustCompile(`<picture>(?:<source type="image/webp" srcset="[^"]+" sizes="([^"]+)">)?<img class="hero__photo" src="/img/profile-720\.[0-9a-f]{8}\.jpg" srcset="[^"]+" sizes="([^"]+)" width="720" height="900" alt="Portrait of Kennedy Muchina Kinuthia" fetchpriority="high">`).FindStringSubmatch(html)
	if hero == nil {
		t.Fatal("hero <picture> missing (720 fallback, fetchpriority=high, not lazy)")
	}
	preload := regexp.MustCompile(`<link rel="preload" as="image"[^>]* imagesrcset="([^"]+)" imagesizes="([^"]+)" fetchpriority="high">`).FindStringSubmatch(html)
	if preload == nil {
		t.Fatal("hero image preload missing")
	}
	if preload[2] != hero[2] {
		t.Errorf("preload imagesizes %q must equal the <img> sizes %q (or the browser fetches twice)", preload[2], hero[2])
	}
	if strings.Contains(preload[1], "%20") {
		t.Errorf("imagesrcset was URL-escaped: %s", preload[1])
	}
	if regexp.MustCompile(`<img [^>]*class="hero__photo"[^>]*loading=`).MatchString(html) {
		t.Error("the hero photo must not be lazy")
	}

	// Head: title, description, canonical, Open Graph, Twitter, author, colour scheme.
	head := html[:strings.Index(html, "</head>")]
	for _, want := range []string{
		`<title>Kennedy Muchina Kinuthia, Software Engineer</title>`,
		`<link rel="canonical" href="https://kmuchina.com/">`,
		`<meta name="author" content="Kennedy Muchina Kinuthia">`,
		`<meta name="color-scheme" content="dark light">`,
		`<meta name="theme-color" content="#0B1220">`,
		`<meta property="og:type" content="profile">`,
		`<meta property="og:url" content="https://kmuchina.com/">`,
		`<meta property="og:image" content="https://kmuchina.com/og.png">`,
		`<meta property="og:image:width" content="1200">`,
		`<meta property="og:image:height" content="630">`,
		`<meta property="og:locale" content="en_KE">`,
		`<meta property="profile:first_name" content="Kennedy">`,
		`<meta property="profile:last_name" content="Kinuthia">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="twitter:image" content="https://kmuchina.com/og.png">`,
		`<link rel="icon" href="/favicon.ico" sizes="32x32">`,
		`<link rel="apple-touch-icon" href="/apple-touch-icon.png">`,
		`<link rel="manifest" href="/site.webmanifest">`,
	} {
		if !strings.Contains(head, want) {
			t.Errorf("<head> missing %s", want)
		}
	}
	desc := regexp.MustCompile(`<meta name="description" content="([^"]+)">`).FindStringSubmatch(head)
	if desc == nil || len([]rune(desc[1])) < 150 || len([]rune(desc[1])) > 160 {
		t.Errorf("meta description should be 150–160 characters: %v", desc)
	}
	// One render-blocking stylesheet, no other blocking script.
	if n := strings.Count(head, `rel="stylesheet"`); n != 1 {
		t.Errorf("want 1 stylesheet, got %d", n)
	}
	if regexp.MustCompile(`<script src="[^"]+"(?:\s*)>`).MatchString(head) {
		t.Error("an external script in <head> without defer blocks rendering")
	}

	// JSON-LD: parse it back and check the Person.
	ld := regexp.MustCompile(`<script type="application/ld\+json">(.+?)</script>`).FindAllStringSubmatch(html, -1)
	if len(ld) != 1 {
		t.Fatalf("want 1 JSON-LD block, got %d", len(ld))
	}
	var p map[string]any
	if err := json.Unmarshal([]byte(ld[0][1]), &p); err != nil {
		t.Fatalf("JSON-LD does not parse: %v", err)
	}
	for key, want := range map[string]string{
		"@context": "https://schema.org", "@type": "Person", "name": "Kennedy Muchina Kinuthia",
		"givenName": "Kennedy", "familyName": "Kinuthia", "jobTitle": "Graduate Software Engineer",
		"url": "https://kmuchina.com/", "email": "mailto:muchina537@gmail.com",
	} {
		if p[key] != want {
			t.Errorf("JSON-LD %s = %v, want %q", key, p[key], want)
		}
	}
	if img, _ := p["image"].(string); !regexp.MustCompile(`^https://kmuchina\.com/img/profile-720\.[0-9a-f]{8}\.jpg$`).MatchString(img) {
		t.Errorf("JSON-LD image = %q, want the absolute 720 JPEG", img)
	} else {
		read(strings.TrimPrefix(img, "https://kmuchina.com/"))
	}
	obj := func(key string) map[string]any { m, _ := p[key].(map[string]any); return m }
	if a := obj("address"); a["@type"] != "PostalAddress" || a["addressLocality"] != "Nairobi" || a["addressCountry"] != "KE" {
		t.Errorf("JSON-LD address = %v", a)
	}
	if a := obj("alumniOf"); a["@type"] != "CollegeOrUniversity" || a["name"] != "Moi University" {
		t.Errorf("JSON-LD alumniOf = %v", a)
	}
	if w := obj("worksFor"); w["@type"] != "Organization" || w["name"] != "Kenya Flower Council" {
		t.Errorf("JSON-LD worksFor = %v", w)
	}
	knows := strings.Join(toStrings(p["knowsAbout"]), "|")
	if !strings.Contains(knows, "Kotlin") || !strings.Contains(knows, "TensorFlow Lite") || strings.Contains(knows, "Git & GitHub") {
		t.Errorf("JSON-LD knowsAbout should be the Programming and Machine learning skills: %s", knows)
	}
	if got := strings.Join(toStrings(p["sameAs"]), " "); got != "https://www.linkedin.com/in/kmuchina https://github.com/kmuchina https://play.google.com/store/apps/developer?id=Merat" {
		t.Errorf("JSON-LD sameAs = %s", got)
	}
	if _, ok := p["telephone"]; ok {
		t.Error("JSON-LD must not carry a telephone")
	}

	// Crawl files.
	if got := read("robots.txt"); got != "User-agent: *\nAllow: /\nDisallow: /styleguide.html\n\nSitemap: https://kmuchina.com/sitemap.xml\n" {
		t.Errorf("robots.txt:\n%s", got)
	}
	var sm struct {
		URLs []struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal([]byte(read("sitemap.xml")), &sm); err != nil {
		t.Fatalf("sitemap.xml: %v", err)
	}
	if len(sm.URLs) != 1 || sm.URLs[0].Loc != "https://kmuchina.com/" || sm.URLs[0].LastMod != time.Now().Format("2006-01-02") {
		t.Errorf("sitemap should list only / with today's lastmod: %+v", sm.URLs)
	}
	if !strings.Contains(read("styleguide.html"), `<meta name="robots" content="noindex">`) {
		t.Error("styleguide.html must stay noindex")
	}

	// Web manifest.
	var mf struct {
		Name, ShortName, Display, ThemeColor, BackgroundColor string
		Icons                                                 []struct{ Src string }
	}
	mfJSON := strings.NewReplacer(`"short_name"`, `"ShortName"`, `"theme_color"`, `"ThemeColor"`, `"background_color"`, `"BackgroundColor"`).Replace(read("site.webmanifest"))
	if err := json.Unmarshal([]byte(mfJSON), &mf); err != nil {
		t.Fatalf("site.webmanifest: %v", err)
	}
	if mf.ShortName != "Kennedy M." || mf.Display != "browser" || mf.ThemeColor != "#0B1220" || mf.BackgroundColor != "#0B1220" || len(mf.Icons) != 2 {
		t.Errorf("site.webmanifest = %+v", mf)
	}
	for _, ic := range mf.Icons {
		read(strings.TrimPrefix(ic.Src, "/"))
	}

	// The stylesheet is minified: no comments, no indentation.
	if m := regexp.MustCompile(`href="/css/(site\.[0-9a-f]{8}\.css)"`).FindStringSubmatch(html); m != nil {
		css := read("css/" + m[1])
		if strings.Contains(css, "/*") || strings.Contains(css, "\n  ") {
			t.Error("site.css should be minified")
		}
	}

	// 404: the spine, its own <h1>, two ways back, the gap trail, noindex.
	nf := read("404.html")
	for _, want := range []string{
		`<meta name="robots" content="noindex">`,
		`<h1 id="not-found-title" class="section__title">Page not found</h1>`,
		`<p class="lead">That page doesn&#39;t exist. It may have moved when the site was rebuilt.</p>`,
		`class="btn btn--primary press" href="/">`,
		`class="btn btn--ghost press" href="/#experience">`,
		`<div class="not-found__trail" aria-hidden="true">`,
	} {
		if !strings.Contains(nf, want) {
			t.Errorf("404.html missing %s", want)
		}
	}
	if strings.Count(nf, "<h1") != 1 || strings.Contains(nf, "application/ld+json") {
		t.Error("404.html: exactly one <h1>, and no JSON-LD")
	}
}

func toStrings(v any) []string {
	var out []string
	if list, ok := v.([]any); ok {
		for _, x := range list {
			s, _ := x.(string)
			out = append(out, s)
		}
	}
	return out
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
