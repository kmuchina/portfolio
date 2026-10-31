package render

import (
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kmuchina/portfolio/internal/assets"
	"github.com/kmuchina/portfolio/internal/images"
	"github.com/kmuchina/portfolio/internal/textfmt"
)

// funcs returns the template FuncMap. Keep it small:
//
//	asset "css/site.css"           → fingerprinted URL path; unknown key = build error
//	icon "github" ["Label"]        → inline SVG (decorative, or labelled with role="img")
//	imgSize "img/profile.jpg"      → assets.Size{W, H} for width/height attributes
//	image "img/profile.jpg"        → *images.Set (srcsets, default <img>); feed it to the "picture" partial
//	imagesrcset set.Srcset         → the whole imagesrcset="…" attribute of an image preload (see below)
//	exists "img/projects/x.png"    → whether the file exists under static/
//	year                           → current year (footer)
//	join list sep                  → strings.Join
//	dict "k1" v1 "k2" v2…          → map[string]any, to pass several values to a partial
//	bold "a **b**"                 → safe HTML: escaped, **b** → <strong>b</strong>
func (r *Renderer) funcs() template.FuncMap {
	return template.FuncMap{
		"asset": func(name string) (string, error) {
			p, ok := r.manifest[name]
			if !ok {
				return "", fmt.Errorf("asset: %q is not in the manifest", name)
			}
			return p, nil
		},
		"icon":    r.icon,
		"imgSize": r.imgSize,
		"image": func(src string) (*images.Set, error) {
			s, ok := r.images[src]
			if !ok {
				return nil, fmt.Errorf("image: %q is not a responsive image (internal/build imageSpecs)", src)
			}
			return s, nil
		},
		"exists": func(rel string) bool {
			_, err := os.Stat(filepath.Join(r.staticDir, filepath.FromSlash(rel)))
			return err == nil
		},
		// html/template reads any attribute with "src" in its name as one
		// URL and would percent-encode a candidate list's spaces and commas;
		// only srcset itself is known to it. The value is our own hashed
		// paths, escaped here.
		"imagesrcset": func(v string) template.HTMLAttr {
			return template.HTMLAttr(`imagesrcset="` + html.EscapeString(v) + `"`)
		},
		"year": func() int { return time.Now().Year() },
		"join": strings.Join,
		"dict": dict,
		"bold": func(s string) template.HTML { return template.HTML(textfmt.Bold(s)) },
	}
}

// icon inlines templates/icons/<name>.svg. With no label the SVG is hidden
// from assistive tech; with a label it gets role="img" and a <title>.
// The SVGs are trusted local files, so returning template.HTML is safe; the
// label is escaped.
func (r *Renderer) icon(name string, label ...string) (template.HTML, error) {
	svg, ok := r.icons[name]
	if !ok {
		return "", fmt.Errorf("icon: no templates/icons/%s.svg", name)
	}
	if len(label) > 1 {
		return "", fmt.Errorf("icon %q: at most one label", name)
	}
	var attrs, title string
	if len(label) == 1 && strings.TrimSpace(label[0]) != "" {
		attrs = ` class="icon" role="img"`
		title = "<title>" + html.EscapeString(label[0]) + "</title>"
	} else {
		attrs = ` class="icon" aria-hidden="true" focusable="false"`
	}
	open := strings.Index(svg, ">")
	out := "<svg" + attrs + svg[len("<svg"):open+1] + title + svg[open+1:]
	return template.HTML(out), nil
}

func (r *Renderer) imgSize(rel string) (assets.Size, error) {
	if s, ok := r.sizes[rel]; ok {
		return s, nil
	}
	s, err := assets.ImageSize(r.staticDir, rel)
	if err != nil {
		return s, fmt.Errorf("imgSize: %w", err)
	}
	r.sizes[rel] = s
	return s, nil
}

// dict builds a map from alternating string keys and values, so a template
// can pass more than one value to a partial:
//
//	{{template "theme-toggle" (dict "Theme" .Site.UI.Theme "Label" .Site.UI.Theme.MenuLabel)}}
func dict(kv ...any) (map[string]any, error) {
	if len(kv)%2 != 0 {
		return nil, fmt.Errorf("dict: odd number of arguments (%d)", len(kv))
	}
	m := make(map[string]any, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: key %d is %T, not a string", i/2, kv[i])
		}
		m[k] = kv[i+1]
	}
	return m, nil
}
