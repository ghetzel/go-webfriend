.PHONY: build ui

PKGS        := $(shell go list ./... 2> /dev/null | grep -v '/vendor')
LOCALS      := $(shell find . -type f -name '*.go' -not -path "./vendor*/*")

.EXPORT_ALL_VARIABLES:
GO111MODULE  = on

all: fmt deps build

fmt:
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	gofmt -w $(LOCALS)
	go build -i -o bin/webfriend-autodoc webfriend/autodoc/*.go
	go generate -x ./...

deps:
	@go list github.com/pointlander/peg || go get github.com/pointlander/peg
	go get ./...
	go vet ./...

test: fmt deps
	go test $(PKGS)

build: fmt
	go build -i -o bin/webfriend webfriend/main.go
	which webfriend && cp -v bin/webfriend `which webfriend` || true
