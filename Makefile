# Portfolio — kmuchina.com
GO ?= go

.PHONY: build serve test contrast vet clean ci check-dist

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

clean: ## Remove build output
	rm -rf dist
