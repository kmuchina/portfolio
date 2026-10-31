package images

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// gradient is an opaque test picture with some detail to encode.
func gradient(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{uint8(x * 255 / w), uint8(y * 255 / h), uint8((x ^ y) & 0xff), 255})
		}
	}
	return img
}

func writeJPEG(t *testing.T, dir, rel string, img image.Image, extra []byte) {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	if extra != nil { // splice a segment in right after SOI
		data = append(append(append([]byte{}, data[:2]...), extra...), data[2:]...)
	}
	writeFile(t, dir, rel, data)
}

func writeFile(t *testing.T, dir, rel string, data []byte) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func widthsOf(vs []Variant) []int {
	var out []int
	for _, v := range vs {
		out = append(out, v.W)
	}
	return out
}

func TestWidths(t *testing.T) {
	for _, c := range []struct {
		want []int
		srcW int
		got  []int
	}{
		{[]int{480, 720, 960, 1200}, 3327, []int{480, 720, 960, 1200}},
		{[]int{480, 720, 960, 1200}, 1000, []int{480, 720, 960}}, // no upscaling
		{[]int{480, 720, 960, 1200}, 960, []int{480, 720, 960}},
		{[]int{360, 540, 720}, 200, []int{200}}, // smaller than every width: source size only
	} {
		if got := Widths(c.want, c.srcW); !reflect.DeepEqual(got, c.got) {
			t.Errorf("Widths(%v, %d) = %v, want %v", c.want, c.srcW, got, c.got)
		}
	}
}

func TestBuildWidthsNamesAndNoWebP(t *testing.T) {
	static, out := t.TempDir(), t.TempDir()
	writeJPEG(t, static, "img/photo.jpg", gradient(1000, 1250), nil)

	sets, err := Build(Options{Static: static, Out: out}, []Spec{{Src: "img/photo.jpg", Widths: []int{480, 720, 960, 1200}, Default: 720}})
	if err != nil {
		t.Fatal(err)
	}
	s := sets["img/photo.jpg"]
	if got := widthsOf(s.Fallback); !reflect.DeepEqual(got, []int{480, 720, 960}) {
		t.Errorf("widths = %v, want 480 720 960 (1200 would upscale)", got)
	}
	if len(s.WebP) != 0 || s.WebPSrcset() != "" {
		t.Error("no cwebp: there should be no WebP variants")
	}
	if s.Type != "image/jpeg" || s.Default.W != 720 || s.Default.H != 900 {
		t.Errorf("default = %s %dx%d, want image/jpeg 720x900", s.Type, s.Default.W, s.Default.H)
	}
	for _, v := range s.Fallback {
		if !strings.HasPrefix(v.URL, "/img/photo-") || !strings.HasSuffix(v.URL, ".jpg") {
			t.Errorf("name %q should be /img/photo-<w>.<hash>.jpg", v.URL)
		}
		f, err := os.Open(filepath.Join(out, filepath.FromSlash(v.URL)))
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := jpeg.DecodeConfig(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Width != v.W || cfg.Height != v.H {
			t.Errorf("%s is %dx%d, recorded as %dx%d", v.URL, cfg.Width, cfg.Height, v.W, v.H)
		}
	}
	if want := s.Fallback[0].URL + " 480w, " + s.Fallback[1].URL + " 720w, " + s.Fallback[2].URL + " 960w"; s.Srcset() != want {
		t.Errorf("srcset = %q", s.Srcset())
	}

	// Same input → same names, with or without the cache.
	cache := t.TempDir()
	for i := range 2 {
		again, err := Build(Options{Static: static, Out: t.TempDir(), Cache: cache}, []Spec{{Src: "img/photo.jpg", Widths: []int{480, 720, 960, 1200}, Default: 720}})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(again["img/photo.jpg"], s) {
			t.Errorf("build %d: names changed:\n%+v\n%+v", i+2, again["img/photo.jpg"], s)
		}
	}
	if files, _ := os.ReadDir(cache); len(files) != 3 {
		t.Errorf("want 3 cached files, got %d", len(files))
	}
}

func TestMetadataStripped(t *testing.T) {
	static, out := t.TempDir(), t.TempDir()
	// An APP1 Exif segment carrying a fake GPS block and camera model.
	payload := []byte("Exif\x00\x00GPSLatitude=1.2921 Model=TestCam")
	seg := append([]byte{0xFF, 0xE1, byte((len(payload) + 2) >> 8), byte(len(payload) + 2)}, payload...)
	writeJPEG(t, static, "img/p.jpg", gradient(600, 750), seg)
	if src, _ := os.ReadFile(filepath.Join(static, "img", "p.jpg")); !bytes.Contains(src, []byte("Exif")) {
		t.Fatal("test setup: the source should carry EXIF")
	}

	sets, err := Build(Options{Static: static, Out: out}, []Spec{{Src: "img/p.jpg", Widths: []int{480}, Default: 480}})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(sets["img/p.jpg"].Default.URL)))
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"Exif", "GPS", "TestCam"} {
		if bytes.Contains(data, []byte(s)) {
			t.Errorf("output still contains %q", s)
		}
	}
}

func TestTransparentPNGStaysPNG(t *testing.T) {
	static := t.TempDir()
	clear := image.NewNRGBA(image.Rect(0, 0, 400, 800)) // fully transparent
	opaque := gradient(400, 800)
	for name, img := range map[string]image.Image{"img/clear.png": clear, "img/opaque.png": opaque} {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		writeFile(t, static, name, buf.Bytes())
	}
	sets, err := Build(Options{Static: static, Out: t.TempDir()}, []Spec{
		{Src: "img/clear.png", Widths: []int{360, 540, 720}, Default: 720},
		{Src: "img/opaque.png", Widths: []int{360, 540, 720}, Default: 720},
	})
	if err != nil {
		t.Fatal(err)
	}
	if s := sets["img/clear.png"]; s.Type != "image/png" || !strings.HasSuffix(s.Default.URL, ".png") {
		t.Errorf("transparent PNG: got %s %s, want PNG", s.Type, s.Default.URL)
	}
	if s := sets["img/opaque.png"]; s.Type != "image/jpeg" {
		t.Errorf("opaque PNG: got %s, want JPEG", s.Type)
	}
	if got := widthsOf(sets["img/clear.png"].Fallback); !reflect.DeepEqual(got, []int{360}) {
		t.Errorf("400px source: widths %v, want just 360", got)
	}
}

// A stand-in cwebp checks the exec path: arguments in, a file out.
func TestWebPWithCwebp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stand-in")
	}
	bin := filepath.Join(t.TempDir(), "cwebp")
	script := "#!/bin/sh\n# args: -quiet -q 80 -m 6 -metadata none in.png -o out.webp\n[ \"$3\" = 80 ] && [ \"$5\" = 6 ] || exit 2\nprintf 'RIFF-fake-webp' > \"${10}\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	static := t.TempDir()
	writeJPEG(t, static, "img/photo.jpg", gradient(800, 1000), nil)
	sets, err := Build(Options{Static: static, Out: t.TempDir(), Cwebp: bin}, []Spec{{Src: "img/photo.jpg", Widths: []int{480, 720}, Default: 720}})
	if err != nil {
		t.Fatal(err)
	}
	s := sets["img/photo.jpg"]
	if got := widthsOf(s.WebP); !reflect.DeepEqual(got, []int{480, 720}) {
		t.Errorf("webp widths = %v", got)
	}
	for _, v := range s.WebP {
		if !strings.HasSuffix(v.URL, ".webp") || v.Bytes != len("RIFF-fake-webp") {
			t.Errorf("webp variant %+v", v)
		}
	}
}
