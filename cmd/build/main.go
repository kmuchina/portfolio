// Command build generates the static site.
//
//	go run ./cmd/build [-content content/site.yaml] [-templates templates] [-static static] [-out dist]
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kmuchina/portfolio/internal/build"
)

func main() {
	o := build.Defaults()
	flag.StringVar(&o.Content, "content", o.Content, "content YAML file")
	flag.StringVar(&o.Templates, "templates", o.Templates, "templates directory")
	flag.StringVar(&o.Static, "static", o.Static, "static assets directory")
	flag.StringVar(&o.Out, "out", o.Out, "output directory (wiped on each build)")
	flag.Parse()

	res, err := build.Run(o)
	for _, w := range res.Warnings {
		if err != nil { // on success the summary prints them
			fmt.Fprintln(os.Stderr, "warning:", w)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "build failed:", err)
		os.Exit(1)
	}
	fmt.Println(res)
}
