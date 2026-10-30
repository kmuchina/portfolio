// Command serve builds the site and serves it for local development.
// Every request for "/" or an .html page triggers a rebuild first, so a
// browser refresh always shows the latest sources.
//
//	go run ./cmd/serve [-addr :8080] [-content …] [-templates …] [-static …] [-out dist]
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/kmuchina/portfolio/internal/build"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	o := build.Defaults()
	flag.StringVar(&o.Content, "content", o.Content, "content YAML file")
	flag.StringVar(&o.Templates, "templates", o.Templates, "templates directory")
	flag.StringVar(&o.Static, "static", o.Static, "static assets directory")
	flag.StringVar(&o.Out, "out", o.Out, "output directory")
	flag.Parse()

	// Rebuilds take the write lock; serving files takes the read lock, so no
	// request ever sees a half-written output directory.
	var mu sync.RWMutex
	rebuild := func() error {
		mu.Lock()
		defer mu.Unlock()
		res, err := build.Run(o)
		if err != nil {
			log.Printf("build failed: %v", err)
			return err
		}
		log.Print(res)
		return nil
	}

	_ = rebuild() // a failing initial build is reported per request below

	files := http.FileServer(http.Dir(o.Out))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") || strings.HasSuffix(p, ".html") {
			if err := rebuild(); err != nil {
				http.Error(w, "build failed:\n\n"+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		mu.RLock()
		defer mu.RUnlock()
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	})

	host, port, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatal(err)
	}
	if host == "" {
		host = "localhost"
	}
	log.Printf("serving %s/ at http://%s:%s/ (styleguide: /styleguide.html)", o.Out, host, port)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
