// Command pngopt recompresses PNG files in place, losslessly, with the
// standard library's best zlib level (headless Chromium writes PNGs fast,
// not small). Fully opaque images lose their alpha channel. `make og` runs
// it on the rendered share image and icons.
//
//	go run ./tools/pngopt static/og.png static/apple-touch-icon.png
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pngopt file.png…")
		os.Exit(2)
	}
	for _, name := range os.Args[1:] {
		if err := optimise(name); err != nil {
			fmt.Fprintf(os.Stderr, "pngopt: %s: %v\n", name, err)
			os.Exit(1)
		}
	}
}

func optimise(name string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	// An opaque RGBA image is written as plain RGB (no alpha bytes).
	if op, ok := img.(interface{ Opaque() bool }); ok && op.Opaque() {
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
		img = rgba
	}
	var buf bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&buf, img); err != nil {
		return err
	}
	if buf.Len() >= len(data) {
		fmt.Printf("%s: %.1f KB (already smaller than a re-encode)\n", name, float64(len(data))/1024)
		return nil
	}
	fmt.Printf("%s: %.1f KB → %.1f KB\n", name, float64(len(data))/1024, float64(buf.Len())/1024)
	return os.WriteFile(name, buf.Bytes(), 0o644)
}
