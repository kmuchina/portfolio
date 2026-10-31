// Package images is the responsive-image pipeline: it resizes each source
// image under static/ to a set of widths (golang.org/x/image/draw,
// CatmullRom), encodes JPEG (opaque sources) or PNG (sources with
// transparency), plus WebP when the cwebp binary is available, and writes
// them under content-hashed names ("img/profile-720.3f9a1c2b.jpg").
//
// Re-encoding drops every metadata segment (EXIF, GPS, XMP, ICC), because
// the encoders only write pixels. Resized outputs are cached on disk, keyed
// by the source bytes and the settings, so repeat builds don't re-encode.
package images

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

// Encoder settings. Changing one changes the cache key (and every name).
const (
	JPEGQuality = 82
	WebPQuality = 80
	version     = "images/v2" // bump to invalidate the cache (v2: cwebp -m 6)
)

// Spec is one source image and the widths to generate from it.
type Spec struct {
	Src     string // path under static/: "img/profile.jpg"
	Widths  []int  // target widths in px; any wider than the source are dropped
	Default int    // width of the <img src> fallback: the largest variant ≤ Default
}

// Variant is one encoded file.
type Variant struct {
	W, H  int
	URL   string // "/img/profile-720.3f9a1c2b.jpg"
	Bytes int
}

// Set is every variant of one source, ready for a <picture>.
type Set struct {
	Src      string
	Type     string    // MIME type of the fallback: "image/jpeg" or "image/png"
	Fallback []Variant // JPEG or PNG, narrowest first
	WebP     []Variant // narrowest first; empty without cwebp
	Default  Variant   // the fallback <img src>
}

// Srcset is the fallback srcset: "/img/a-480.x.jpg 480w, …".
func (s *Set) Srcset() string { return srcset(s.Fallback) }

// WebPSrcset is the WebP srcset, or "" when there is no WebP.
func (s *Set) WebPSrcset() string { return srcset(s.WebP) }

func srcset(vs []Variant) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = v.URL + " " + strconv.Itoa(v.W) + "w"
	}
	return strings.Join(parts, ", ")
}

// Options configures Build.
type Options struct {
	Static string // source directory
	Out    string // output directory (already created)
	Cache  string // cache directory; "" disables the cache
	Cwebp  string // path of the cwebp binary; "" skips WebP
}

// LookCwebp returns the path of cwebp on PATH, or "".
func LookCwebp() string {
	p, err := exec.LookPath("cwebp")
	if err != nil {
		return ""
	}
	return p
}

// Build generates every spec's variants into o.Out and returns the sets,
// keyed by Spec.Src.
func Build(o Options, specs []Spec) (map[string]*Set, error) {
	sets := make(map[string]*Set, len(specs))
	for _, sp := range specs {
		s, err := build(o, sp)
		if err != nil {
			return nil, fmt.Errorf("images: %s: %w", sp.Src, err)
		}
		sets[sp.Src] = s
	}
	return sets, nil
}

// Widths returns the widths actually generated for a source srcW px wide:
// the requested ones that don't upscale, or just srcW when all would.
func Widths(want []int, srcW int) []int {
	var out []int
	for _, w := range want {
		if w > 0 && w <= srcW {
			out = append(out, w)
		}
	}
	if len(out) == 0 {
		out = []int{srcW}
	}
	return out
}

// source lazily decodes the image the first time a variant isn't cached.
type source struct {
	path   string
	data   []byte
	sum    string // hex sha256 of data
	w, h   int
	format string // "jpeg" | "png"

	once  sync.Once
	img   image.Image
	alpha bool
	err   error
}

func (s *source) decode() (image.Image, error) {
	s.once.Do(func() {
		s.img, _, s.err = image.Decode(bytes.NewReader(s.data))
	})
	return s.img, s.err
}

func build(o Options, sp Spec) (*Set, error) {
	src := &source{path: filepath.Join(o.Static, filepath.FromSlash(sp.Src))}
	var err error
	if src.data, err = os.ReadFile(src.path); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(src.data)
	src.sum = hex.EncodeToString(sum[:])
	cfg, format, err := image.DecodeConfig(bytes.NewReader(src.data))
	if err != nil {
		return nil, err
	}
	src.w, src.h, src.format = cfg.Width, cfg.Height, format

	// Opaque sources (every JPEG, and PNGs without transparency) become
	// JPEG; a PNG with any transparent pixel stays PNG.
	fallback := "jpeg"
	if format == "png" {
		alpha, err := hasAlpha(o, src)
		if err != nil {
			return nil, err
		}
		if alpha {
			fallback = "png"
		}
	}

	widths := Widths(sp.Widths, src.w)
	set := &Set{Src: sp.Src, Type: "image/" + fallback}
	formats := []string{fallback}
	if o.Cwebp != "" {
		formats = append(formats, "webp")
	}

	type job struct {
		format string
		w, i   int
	}
	results := make(map[string][]Variant, len(formats))
	for _, f := range formats {
		results[f] = make([]Variant, len(widths))
	}
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	for _, f := range formats {
		for i, w := range widths {
			wg.Add(1)
			go func(j job) {
				defer wg.Done()
				v, err := variant(o, src, sp.Src, j.format, j.w)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errs = append(errs, fmt.Errorf("%s %dw: %w", j.format, j.w, err))
					return
				}
				results[j.format][j.i] = v
			}(job{f, w, i})
		}
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	set.Fallback = results[fallback]
	set.WebP = results["webp"]
	set.Default = set.Fallback[0]
	for _, v := range set.Fallback {
		if v.W <= sp.Default {
			set.Default = v
		}
	}
	return set, nil
}

// variant returns one encoded width, from the cache when possible, and
// writes it to the output under its content-hashed name.
func variant(o Options, src *source, rel, format string, w int) (Variant, error) {
	h := max(1, (src.h*w+src.w/2)/src.w) // keep the aspect ratio, rounded
	q := map[string]int{"jpeg": JPEGQuality, "webp": WebPQuality, "png": 0}[format]
	key := cacheKey(src.sum, format, w, h, q)
	ext := map[string]string{"jpeg": ".jpg", "png": ".png", "webp": ".webp"}[format]

	var data []byte
	cached := ""
	if o.Cache != "" {
		cached = filepath.Join(o.Cache, key+ext)
		data, _ = os.ReadFile(cached)
	}
	if data == nil {
		img, err := src.decode()
		if err != nil {
			return Variant{}, err
		}
		dst := resize(img, w, h)
		if data, err = encode(o, dst, format); err != nil {
			return Variant{}, err
		}
		if cached != "" {
			if err := os.MkdirAll(o.Cache, 0o755); err != nil {
				return Variant{}, err
			}
			if err := writeAtomic(cached, data); err != nil {
				return Variant{}, err
			}
		}
	}

	sum := sha256.Sum256(data)
	base := strings.TrimSuffix(rel, path.Ext(rel))
	out := fmt.Sprintf("%s-%d.%s%s", base, w, hex.EncodeToString(sum[:])[:8], ext)
	dst := filepath.Join(o.Out, filepath.FromSlash(out))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return Variant{}, err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return Variant{}, err
	}
	return Variant{W: w, H: h, URL: "/" + out, Bytes: len(data)}, nil
}

func cacheKey(srcSum, format string, w, h, q int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%dx%d|q%d", version, srcSum, format, w, h, q)))
	return hex.EncodeToString(sum[:])[:32]
}

// resize scales img to w × h with Catmull-Rom (sharp, no ringing to speak
// of when downscaling photos).
func resize(img image.Image, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	return dst
}

func encode(o Options, img *image.NRGBA, format string) ([]byte, error) {
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality})
		return buf.Bytes(), err
	case "png":
		err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&buf, img)
		return buf.Bytes(), err
	case "webp":
		return cwebp(o.Cwebp, img)
	}
	return nil, fmt.Errorf("unknown format %q", format)
}

// cwebp encodes img with the cwebp binary: a lossless PNG goes in, a
// metadata-free WebP at WebPQuality comes out.
func cwebp(bin string, img *image.NRGBA) ([]byte, error) {
	dir, err := os.MkdirTemp("", "cwebp-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	in, out := filepath.Join(dir, "in.png"), filepath.Join(dir, "out.webp")
	var buf bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&buf, img); err != nil {
		return nil, err
	}
	if err := os.WriteFile(in, buf.Bytes(), 0o644); err != nil {
		return nil, err
	}
	// -m 6: cwebp's slowest, strongest compression (same quality, smaller files).
	cmd := exec.Command(bin, "-quiet", "-q", strconv.Itoa(WebPQuality), "-m", "6", "-metadata", "none", in, "-o", out)
	if msg, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("cwebp: %v: %s", err, bytes.TrimSpace(msg))
	}
	return os.ReadFile(out)
}

// hasAlpha reports whether a PNG has any non-opaque pixel. It decodes the
// image, so the result is worth caching too: a marker file per source.
func hasAlpha(o Options, src *source) (bool, error) {
	marker := ""
	if o.Cache != "" {
		marker = filepath.Join(o.Cache, cacheKey(src.sum, "alpha", 0, 0, 0))
		if b, err := os.ReadFile(marker); err == nil {
			return string(b) == "1", nil
		}
	}
	img, err := src.decode()
	if err != nil {
		return false, err
	}
	alpha := true
	if op, ok := img.(interface{ Opaque() bool }); ok {
		alpha = !op.Opaque()
	}
	if marker != "" {
		if err := os.MkdirAll(o.Cache, 0o755); err != nil {
			return false, err
		}
		v := "0"
		if alpha {
			v = "1"
		}
		if err := writeAtomic(marker, []byte(v)); err != nil {
			return false, err
		}
	}
	return alpha, nil
}

// writeAtomic writes via a temp file and rename, so an interrupted build
// never leaves a truncated file in the cache.
func writeAtomic(name string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(name), ".tmp-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), name)
}
