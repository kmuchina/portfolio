// Package assets is the static pipeline: copy static/ to the output,
// concatenate, minify (MinifyCSS) and fingerprint CSS, fingerprint main.js,
// and build the manifest that the template `asset` func reads.
package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // register decoders for DecodeConfig
	_ "image/png"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// Manifest maps logical static paths ("css/site.css") to the URL paths
// written to the output ("/css/site.3f9a1c2b.css"). Every copied file is in
// it, so templates can reference any asset through `asset`.
type Manifest map[string]string

// Result summarises the static pipeline.
type Result struct {
	Manifest      Manifest
	Files         int
	Bytes         int64
	Fingerprinted []string // output paths of fingerprinted files
}

// Files that are never copied as-is.
const (
	cssDir      = "css"
	mainJS      = "js/main.js"
	themeInitJS = "js/theme-init.js" // inlined into <head> by the layout
)

// Build wipes outDir and writes the static assets into it. Files in skip
// (static-relative paths, e.g. the sources of the responsive images) are
// not copied.
func Build(staticDir, outDir string, skip []string, protect ...string) (*Result, error) {
	if err := checkOut(outDir, append(protect, staticDir)...); err != nil {
		return nil, err
	}
	if err := os.RemoveAll(outDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}

	res := &Result{Manifest: Manifest{}}
	var cssFiles []string

	err := filepath.WalkDir(staticDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(staticDir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(d.Name(), ".") && rel != "." { // .gitkeep, .DS_Store
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		switch {
		case slices.Contains(skip, rel):
			return nil
		case path.Dir(rel) == cssDir && path.Ext(rel) == ".css":
			cssFiles = append(cssFiles, p)
			return nil
		case rel == themeInitJS:
			return nil
		case rel == mainJS:
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			return res.writeHashed(outDir, mainJS, data)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return res.write(outDir, rel, rel, data)
	})
	if err != nil {
		return nil, err
	}

	if len(cssFiles) == 0 {
		return nil, fmt.Errorf("no CSS files in %s/%s", staticDir, cssDir)
	}
	sort.Strings(cssFiles)
	var css bytes.Buffer
	for _, f := range cssFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		css.Write(bytes.TrimSpace(data))
		css.WriteString("\n")
	}
	if err := res.writeHashed(outDir, "css/site.css", MinifyCSS(css.Bytes())); err != nil {
		return nil, err
	}
	return res, nil
}

// write copies data to outDir/out and records logical → /out.
func (r *Result) write(outDir, logical, out string, data []byte) error {
	dst := filepath.Join(outDir, filepath.FromSlash(out))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	r.Manifest[logical] = "/" + out
	r.Files++
	r.Bytes += int64(len(data))
	return nil
}

// writeHashed writes "dir/name.<hash8>.ext" for logical "dir/name.ext".
func (r *Result) writeHashed(outDir, logical string, data []byte) error {
	sum := sha256.Sum256(data)
	ext := path.Ext(logical)
	out := strings.TrimSuffix(logical, ext) + "." + hex.EncodeToString(sum[:])[:8] + ext
	r.Fingerprinted = append(r.Fingerprinted, out)
	return r.write(outDir, logical, out, data)
}

// Size is an image's pixel dimensions.
type Size struct{ W, H int }

// ImageSize reads the dimensions of a JPEG or PNG under staticDir without
// decoding the whole image.
func ImageSize(staticDir, rel string) (Size, error) {
	f, err := os.Open(filepath.Join(staticDir, filepath.FromSlash(rel)))
	if err != nil {
		return Size{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return Size{}, fmt.Errorf("%s: %w", rel, err)
	}
	return Size{W: cfg.Width, H: cfg.Height}, nil
}

// checkOut refuses to wipe a directory that is, or contains, a source path.
func checkOut(out string, protect ...string) error {
	absOut, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if out == "" || absOut == filepath.Dir(absOut) {
		return errors.New("refusing to use an empty or root output directory")
	}
	for _, p := range protect {
		absP, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		if absP == absOut || strings.HasPrefix(absP, absOut+string(filepath.Separator)) {
			return fmt.Errorf("refusing to wipe %s: it contains %s", out, p)
		}
	}
	return nil
}

var (
	tokenCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	tokenBlockRe   = regexp.MustCompile(`([^{}]+)\{([^{}]*)\}`)
)

// DarkColor reads a colour token of the dark (default) theme from
// static/css/00-tokens.css, e.g. DarkColor(static, "bg") = "#0B1220", so
// the outputs that can't read CSS variables (the theme-color meta, the web
// manifest) never carry a second copy of a hex value.
func DarkColor(staticDir, name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(staticDir, "css", "00-tokens.css"))
	if err != nil {
		return "", err
	}
	css := tokenCommentRe.ReplaceAllString(string(data), "")
	valRe := regexp.MustCompile(`--` + regexp.QuoteMeta(name) + `\s*:\s*(#[0-9a-fA-F]{6})\s*;`)
	for _, b := range tokenBlockRe.FindAllStringSubmatch(css, -1) {
		if !strings.Contains(b[1], `[data-theme="dark"]`) {
			continue
		}
		if m := valRe.FindStringSubmatch(b[2]); m != nil {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("00-tokens.css: no dark-theme --%s hex value", name)
}
