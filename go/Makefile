# git-context Makefile
BINARY=git-ctx
VERSION=0.2.0
GOFLAGS=-ldflags="-s -w"

.PHONY: build install test clean release

build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/git-ctx

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

# Cross-platform release builds
release: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/git-ctx
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/git-ctx
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/git-ctx
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o dist/$(BINARY)-linux-arm64 ./cmd/git-ctx
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/git-ctx
	@echo "Release builds in dist/"

