.PHONY: build ui

PKGS=`go list ./... 2> /dev/null | grep -v '/vendor'`
LOCALS=`find . -type f -name '*.go' -not -path "./vendor*/*"`


all: fmt deps build

fmt:
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	goimports -w $(LOCALS)
	go vet .

deps:
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	@go list github.com/pointlander/peg || go get github.com/pointlander/peg
	go generate -x ./scripting
	go get .

test: fmt deps
	go test $(PKGS)

build: fmt
	go build -i -o bin/webfriend webfriend/main.go
	go build -i -o bin/webfriend-autodoc webfriend/autodoc/*.go
