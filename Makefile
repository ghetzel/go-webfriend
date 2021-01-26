.PHONY: build ui docs

LOCALS         := $(shell find . -type f -name '*.go' -not -path "./vendor*/*")
WEBFRIEND_BIN   = webfriend-$(shell go env GOOS)-$(shell go env GOARCH)
BIN_VERSION     = $(shell ./bin/$(WEBFRIEND_BIN) --version | cut -d' ' -f3)
CGO_ENABLED    ?= 0

.EXPORT_ALL_VARIABLES:
GO111MODULE  = on

all: fmt deps autodoc build docs

fmt:
	@go list github.com/mjibson/esc || go get github.com/mjibson/esc/...
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	gofmt -w $(LOCALS)
	go vet ./...
	go mod tidy

deps:
	@go list github.com/pointlander/peg || go get github.com/pointlander/peg
	go get ./...

test: fmt deps
	go test ./...

autodoc:
	go build -o bin/webfriend-autodoc cmd/webfriend-autodoc/*.go
	go generate -x ./...

docs: fmt
	cd docs && make

$(WEBFRIEND_BIN):
	go build -tags nocgo --ldflags '-extldflags "-static"' -ldflags '-s' -o bin/$(WEBFRIEND_BIN) cmd/webfriend/*.go
	GOARCH=amd64 go build -tags nocgo --ldflags '-extldflags "-static"' -ldflags '-s' -o bin/webfriend-$(shell go env GOOS)-amd64 cmd/webfriend/*.go
	which webfriend && cp -v bin/$(WEBFRIEND_BIN) `which webfriend` || true

build: $(WEBFRIEND_BIN)

docker:
	docker build -t ghetzel/webfriend:$(BIN_VERSION) .

docker-push:
	docker tag ghetzel/webfriend:$(BIN_VERSION) ghetzel/webfriend:latest
	docker push ghetzel/webfriend:$(BIN_VERSION)
	docker push ghetzel/webfriend:latest