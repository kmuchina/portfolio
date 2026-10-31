// Package render parses the templates and executes every page in
// templates/pages/ into the output directory.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kmuchina/portfolio/internal/assets"
	"github.com/kmuchina/portfolio/internal/career"
	"github.com/kmuchina/portfolio/internal/evidence"
	"github.com/kmuchina/portfolio/internal/images"
	"github.com/kmuchina/portfolio/internal/seo"
	"github.com/kmuchina/portfolio/internal/site"
)

// Page is the data passed to every template.
type Page struct {
	Site      *site.Site
	Name      string           // page file name without .html ("index", "404")
	Path      string           // URL path ("/", "/404.html")
	Home      string           // prefix for in-page anchors: "" on index, "/" elsewhere
	Canonical string           // absolute URL; empty when NoIndex
	NoIndex   bool             // page template defines {{define "noindex"}}{{end}}
	ThemeInit template.JS      // contents of static/js/theme-init.js, inlined in <head>
	Career    *career.Graph    // the experience graph, built from site.yaml for the current month
	Skills    []evidence.Group // skill groups with each skill's evidence (internal/evidence)
	LocalTime LocalTime        // build-time clock in contact.timezone (main.js takes over)

	ThemeColor string      // the dark --bg, read from 00-tokens.css (main.js re-syncs it)
	ShareImage string      // absolute URL of seo.image (og:image)
	JSONLD     template.JS // the Person JSON-LD (encoding/json); indexable pages only
}

// LocalTime is the contact section's no-JS clock: the time of the build in
// contact.timezone.
type LocalTime struct {
	Text string // "15:26"
	ISO  string // "2026-10-30T15:26+03:00"
}

// Renderer holds everything the templates need.
type Renderer struct {
	templatesDir string
	staticDir    string
	manifest     assets.Manifest
	icons        map[string]string
	sizes        map[string]assets.Size
	images       map[string]*images.Set
}

// New loads the icon SVGs and prepares a renderer. imgs are the responsive
// image sets (internal/images), keyed by their static path.
func New(templatesDir, staticDir string, manifest assets.Manifest, imgs map[string]*images.Set) (*Renderer, error) {
	r := &Renderer{
		templatesDir: templatesDir,
		staticDir:    staticDir,
		manifest:     manifest,
		icons:        map[string]string{},
		sizes:        map[string]assets.Size{},
		images:       imgs,
	}
	files, err := filepath.Glob(filepath.Join(templatesDir, "icons", "*.svg"))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		svg := strings.TrimSpace(string(data))
		if !strings.HasPrefix(svg, "<svg") {
			return nil, fmt.Errorf("%s: must start with <svg", f)
		}
		r.icons[strings.TrimSuffix(filepath.Base(f), ".svg")] = svg
	}
	return r, nil
}

// Output describes one rendered page.
type Output struct {
	File  string
	Bytes int
}

// Render executes every templates/pages/*.html with the layout and partials
// and writes outDir/<name>.html.
func (r *Renderer) Render(s *site.Site, outDir string) ([]Output, error) {
	themeInit, err := os.ReadFile(filepath.Join(r.staticDir, "js", "theme-init.js"))
	if err != nil {
		return nil, err
	}

	base, err := template.New("").Funcs(r.funcs()).ParseFiles(filepath.Join(r.templatesDir, "layout.html"))
	if err != nil {
		return nil, err
	}
	if _, err := base.ParseGlob(filepath.Join(r.templatesDir, "partials", "*.html")); err != nil {
		return nil, err
	}

	pages, err := filepath.Glob(filepath.Join(r.templatesDir, "pages", "*.html"))
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages in %s/pages", r.templatesDir)
	}
	sort.Strings(pages)

	now := time.Now()
	graph, err := careerGraph(s, now)
	if err != nil {
		return nil, err
	}
	ui := s.UI.Skills
	skills := evidence.Build(s, evidence.Labels{
		UsedIn: ui.UsedIn, UsedAt: ui.UsedAt, ShownAt: ui.ShownAt, Project: ui.Project,
		Place: ui.Place, Places: ui.Places,
	})
	loc, err := time.LoadLocation(s.Contact.Timezone.IANA)
	if err != nil {
		return nil, fmt.Errorf("contact.timezone.iana: %w", err)
	}
	local := now.In(loc)
	clock := LocalTime{Text: local.Format("15:04"), ISO: local.Format("2006-01-02T15:04-07:00")}

	themeColor, err := assets.DarkColor(r.staticDir, "bg")
	if err != nil {
		return nil, err
	}
	shareImage, ok := r.manifest[s.SEO.Image]
	if !ok {
		return nil, fmt.Errorf("seo.image %q is not in the asset manifest", s.SEO.Image)
	}
	photo, ok := r.images[s.Person.Photo]
	if !ok {
		return nil, fmt.Errorf("person.photo %q has no responsive images", s.Person.Photo)
	}
	jsonld, err := seo.JSONLD(seo.NewPerson(s, seo.Abs(s.Meta.URL, photo.Default.URL)))
	if err != nil {
		return nil, err
	}

	var out []Output
	for _, file := range pages {
		t, err := base.Clone()
		if err != nil {
			return nil, err
		}
		if _, err := t.ParseFiles(file); err != nil {
			return nil, err
		}

		name := strings.TrimSuffix(filepath.Base(file), ".html")
		p := Page{
			Site:      s,
			Name:      name,
			Path:      "/" + name + ".html",
			Home:      "/",
			NoIndex:   t.Lookup("noindex") != nil,
			ThemeInit: template.JS(bytes.TrimSpace(themeInit)),
			Career:    graph,
			Skills:    skills,
			LocalTime: clock,

			ThemeColor: themeColor,
			ShareImage: seo.Abs(s.Meta.URL, shareImage),
		}
		if name == "index" {
			p.Path, p.Home = "/", ""
		}
		if !p.NoIndex {
			p.Canonical = strings.TrimRight(s.Meta.URL, "/") + p.Path
			p.JSONLD = template.JS(jsonld)
		}

		var buf bytes.Buffer
		if err := t.ExecuteTemplate(&buf, "layout", p); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		dst := filepath.Join(outDir, name+".html")
		if err := os.WriteFile(dst, buf.Bytes(), 0o644); err != nil {
			return nil, err
		}
		out = append(out, Output{File: name + ".html", Bytes: buf.Len()})
	}
	return out, nil
}
