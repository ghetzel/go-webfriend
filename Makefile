.PHONY: build ui docs

WEBFRIEND_BIN   = webfriend-$(shell go env GOOS)-$(shell go env GOARCH)

.EXPORT_ALL_VARIABLES:

all: fmt deps build

fmt:
	@go fmt .
	@go vet ./...
	@go mod tidy

deps:
	@go get ./...

test: fmt deps
	@go test ./...

autodoc:
	@go build -o bin/webfriend-autodoc cmd/webfriend-autodoc/*.go
	@go generate -x ./...

$(WEBFRIEND_BIN):
	go build -o bin/$(WEBFRIEND_BIN) cmd/webfriend/*.go
	GOARCH=amd64 go build -tags nocgo --ldflags '-extldflags "-static"' -ldflags '-s' -o bin/webfriend-$(shell go env GOOS)-amd64 cmd/webfriend/*.go
#	@which webfriend && cp -v bin/$(WEBFRIEND_BIN) `which webfriend` || true

build: $(WEBFRIEND_BIN)

.PHONY: build $(WEBFRIEND_BIN)
