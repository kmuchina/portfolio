// Package build orchestrates a full site build:
// load content → validate → static assets → render pages.
// Shared by cmd/build and cmd/serve.
package build

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kmuchina/portfolio/internal/assets"
	"github.com/kmuchina/portfolio/internal/images"
	"github.com/kmuchina/portfolio/internal/render"
	"github.com/kmuchina/portfolio/internal/seo"
	"github.com/kmuchina/portfolio/internal/site"
)

// Responsive image widths (internal/images). The templates pass each use's
// `sizes`; these say which files exist.
var (
	PhotoWidths      = []int{480, 720, 960, 1200} // person.photo (hero)
	ScreenshotWidths = []int{360, 540, 720}       // projects[].screenshot (phone screens)
)

// fallbackWidth is the <img src> of every responsive image (the largest
// variant at most this wide).
const fallbackWidth = 720

// Options configures a build. Paths are relative to the working directory.
type Options struct {
	Content   string // content YAML (default "content/site.yaml")
	Templates string // templates directory (default "templates")
	Static    string // static directory (default "static")
	Out       string // output directory, wiped on each build (default "dist")
	Cache     string // resized-image cache; "" = none (default ".cache/images")
	Cwebp     string // cwebp binary for WebP images; "" = no WebP (default: cwebp on PATH)
}

// Defaults returns the repo-root defaults.
func Defaults() Options {
	return Options{
		Content:   "content/site.yaml",
		Templates: "templates",
		Static:    "static",
		Out:       "dist",
		Cache:     ".cache/images",
		Cwebp:     images.LookCwebp(),
	}
}

// Result summarises a build.
type Result struct {
	Out           string
	Files         int
	Bytes         int64
	Pages         []string
	Fingerprinted []string
	Warnings      []string
	Took          time.Duration
}

func (r Result) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "built %s/: %d files, %.1f KB in %s · pages: %s · %s",
		r.Out, r.Files, float64(r.Bytes)/1024, r.Took.Round(time.Millisecond),
		strings.Join(r.Pages, ", "), strings.Join(r.Fingerprinted, ", "))
	for _, w := range r.Warnings {
		b.WriteString("\n  warning: " + w)
	}
	return b.String()
}

// Run performs a full build.
func Run(o Options) (Result, error) {
	start := time.Now()
	res := Result{Out: filepath.ToSlash(filepath.Clean(o.Out))}

	s, err := site.Load(o.Content)
	if err != nil {
		return res, err
	}
	warnings, err := site.Validate(s, o.Static, filepath.Join(o.Templates, "icons"))
	res.Warnings = warnings
	if err != nil {
		return res, err
	}

	// The image sources are never copied as-is: only their resized variants.
	specs := imageSpecs(s, o.Static)
	var sources []string
	for _, sp := range specs {
		sources = append(sources, sp.Src)
	}
	protect := []string{o.Content, o.Templates}
	if o.Cache != "" {
		protect = append(protect, o.Cache)
	}
	a, err := assets.Build(o.Static, o.Out, sources, protect...)
	if err != nil {
		return res, err
	}

	if o.Cwebp == "" {
		res.Warnings = append(res.Warnings, "cwebp not found on PATH: WebP images skipped (install the webp package)")
	}
	imgs, err := images.Build(images.Options{Static: o.Static, Out: o.Out, Cache: o.Cache, Cwebp: o.Cwebp}, specs)
	if err != nil {
		return res, err
	}
	for _, set := range imgs {
		for _, v := range append(set.Fallback, set.WebP...) {
			a.Files++
			a.Bytes += int64(v.Bytes)
		}
	}

	if err := writeWebManifest(s, o.Static, o.Out, a.Manifest); err != nil {
		return res, err
	}
	a.Files++

	r, err := render.New(o.Templates, o.Static, a.Manifest, imgs)
	if err != nil {
		return res, err
	}
	pages, err := r.Render(s, o.Out)
	if err != nil {
		return res, err
	}

	pagesFiles, err := writePagesFiles(s, o.Out)
	if err != nil {
		return res, err
	}
	crawlFiles, err := writeCrawlFiles(s, o.Out, start)
	if err != nil {
		return res, err
	}

	res.Files = a.Files + pagesFiles + crawlFiles
	res.Bytes = a.Bytes
	res.Fingerprinted = a.Fingerprinted
	for _, p := range pages {
		res.Files++
		res.Bytes += int64(p.Bytes)
		res.Pages = append(res.Pages, p.File)
	}
	res.Took = time.Since(start)
	return res, nil
}

// writePagesFiles writes the GitHub Pages control files into out:
// CNAME (the custom domain, the host of site.url) and an empty .nojekyll
// (serve the output as-is, without Jekyll). It returns the number written.
func writePagesFiles(s *site.Site, out string) (int, error) {
	u, err := url.Parse(s.Meta.URL)
	if err != nil || u.Hostname() == "" {
		return 0, fmt.Errorf("site.url %q: no host for CNAME", s.Meta.URL)
	}
	files := map[string]string{
		"CNAME":     u.Hostname() + "\n",
		".nojekyll": "",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(out, name), []byte(body), 0o644); err != nil {
			return 0, err
		}
	}
	return len(files), nil
}

// imageSpecs lists the responsive images: the profile photo, and every
// project screenshot that exists (a missing one is a validation warning).
func imageSpecs(s *site.Site, static string) []images.Spec {
	specs := []images.Spec{{Src: s.Person.Photo, Widths: PhotoWidths, Default: fallbackWidth}}
	for _, p := range s.Projects.Items {
		if p.Screenshot == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(static, filepath.FromSlash(p.Screenshot))); err == nil {
			specs = append(specs, images.Spec{Src: p.Screenshot, Widths: ScreenshotWidths, Default: fallbackWidth})
		}
	}
	return specs
}

// writeCrawlFiles writes robots.txt and sitemap.xml (only the home page,
// lastmod = the build date). It returns the number written.
func writeCrawlFiles(s *site.Site, out string, now time.Time) (int, error) {
	sitemap, err := seo.Sitemap(s.Meta.URL, now)
	if err != nil {
		return 0, err
	}
	files := map[string][]byte{
		"robots.txt":  seo.Robots(s.Meta.URL),
		"sitemap.xml": sitemap,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(out, name), body, 0o644); err != nil {
			return 0, err
		}
	}
	return len(files), nil
}

// writeWebManifest writes site.webmanifest (name, icons, the dark --bg as
// both colours; no service worker) and adds it to the asset manifest.
func writeWebManifest(s *site.Site, static, out string, m assets.Manifest) error {
	bg, err := assets.DarkColor(static, "bg")
	if err != nil {
		return err
	}
	var icons []seo.Icon
	for _, ic := range []seo.Icon{
		{Src: "favicon.svg", Sizes: "any", Type: "image/svg+xml"},
		{Src: "apple-touch-icon.png", Sizes: "180x180", Type: "image/png"},
	} {
		url, ok := m[ic.Src]
		if !ok {
			return fmt.Errorf("site.webmanifest: static/%s is missing", ic.Src)
		}
		ic.Src = url
		icons = append(icons, ic)
	}
	body, err := seo.Manifest{
		Name:            s.Person.FullName(),
		ShortName:       s.SEO.ShortName,
		StartURL:        "/",
		Display:         "browser",
		ThemeColor:      bg,
		BackgroundColor: bg,
		Icons:           icons,
	}.JSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "site.webmanifest"), body, 0o644); err != nil {
		return err
	}
	m["site.webmanifest"] = "/site.webmanifest"
	return nil
}
