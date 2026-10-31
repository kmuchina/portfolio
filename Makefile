# Portfolio — kmuchina.com
GO ?= go

.PHONY: build serve test contrast vet clean ci check-dist og

build: ## Build the static site into dist/
	$(GO) run ./cmd/build

serve: ## Build and serve on http://localhost:8080 (rebuilds on page refresh)
	$(GO) run ./cmd/serve

test: ## Run the Go tests (content validation + full build)
	$(GO) test ./...

contrast: ## Check WCAG contrast of the colour tokens (both themes)
	$(GO) run ./tools/contrast

vet: ## Static checks
	$(GO) vet ./...

ci: vet test build check-dist ## Run exactly what CI runs before deploying

check-dist: ## Check the build output is ready for GitHub Pages
	test -f dist/index.html
	test -f dist/404.html
	test -f dist/.nojekyll
	grep -qx 'kmuchina.com' dist/CNAME
	for f in robots.txt sitemap.xml og.png favicon.svg favicon.ico apple-touch-icon.png site.webmanifest; do test -f "dist/$$f"; done
	test ! -e dist/img/profile.jpg

og: ## Re-render static/og.png, favicon.ico and apple-touch-icon.png from design/ (needs Playwright)
	@if command -v node >/dev/null 2>&1; then node tools/og/render.cjs && \
	  $(GO) run ./tools/pngopt static/og.png static/apple-touch-icon.png; \
	else echo "make og needs Node and Playwright: npm install -g playwright && npx playwright install chromium"; exit 1; fi

clean: ## Remove build output
	rm -rf dist
