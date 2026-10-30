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
	"github.com/kmuchina/portfolio/internal/render"
	"github.com/kmuchina/portfolio/internal/site"
)

// Options configures a build. Paths are relative to the working directory.
type Options struct {
	Content   string // content YAML (default "content/site.yaml")
	Templates string // templates directory (default "templates")
	Static    string // static directory (default "static")
	Out       string // output directory, wiped on each build (default "dist")
}

// Defaults returns the repo-root defaults.
func Defaults() Options {
	return Options{
		Content:   "content/site.yaml",
		Templates: "templates",
		Static:    "static",
		Out:       "dist",
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

	a, err := assets.Build(o.Static, o.Out, o.Content, o.Templates)
	if err != nil {
		return res, err
	}

	r, err := render.New(o.Templates, o.Static, a.Manifest)
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

	res.Files = a.Files + pagesFiles
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
