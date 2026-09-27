# git-context Makefile
BINARY=git-ctx
VERSION=0.3.0
MODULE=github.com/jxucoder/git-context
LDFLAGS=-ldflags="-s -w -X $(MODULE)/internal/cmd.version=$(VERSION)"

.PHONY: build install test clean deps release

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/git-ctx

install: build
	mkdir -p ~/.local/bin
	cp $(BINARY) ~/.local/bin/
	git config --global alias.ctx '!git-ctx'
	@echo "Installed git-ctx to ~/.local/bin/"
	@echo "Run 'git ctx --help' to get started"

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

deps:
	go mod tidy
	go mod download

# Cross-platform release builds (CGO disabled for cross-compilation)
release: clean
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/git-ctx
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/git-ctx
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/git-ctx
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 ./cmd/git-ctx
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/git-ctx
	cd dist && shasum -a 256 $(BINARY)-* > checksums.txt
	@echo "Release builds in dist/"
