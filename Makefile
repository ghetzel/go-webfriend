.PHONY: build ui docs

PKGS           := $(shell go list ./... 2> /dev/null | grep -v '/vendor')
LOCALS         := $(shell find . -type f -name '*.go' -not -path "./vendor*/*")
WEBFRIEND_BIN  ?= webfriend-$(shell go env GOOS)-$(shell go env GOARCH)

.EXPORT_ALL_VARIABLES:
GO111MODULE  = on

all: fmt deps build docs

fmt:
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	gofmt -w $(LOCALS)
	go build -i -o bin/webfriend-autodoc cmd/webfriend-autodoc/*.go
	go generate -x ./...

deps:
	@go list github.com/pointlander/peg || go get github.com/pointlander/peg
	go get ./...
	-go mod tidy
	go vet ./...

test: fmt deps
	go test $(PKGS)

docs: fmt
	cd docs && make

build: fmt
	go build -tags nocgo --ldflags '-extldflags "-static"' -ldflags '-s' -o bin/$(WEBFRIEND_BIN) cmd/webfriend/*.go
	which webfriend && cp -v bin/$(WEBFRIEND_BIN) `which webfriend` || true
