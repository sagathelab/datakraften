BINARY=dk
GO=$(shell which go 2>/dev/null || echo "/tmp/go/bin/go")
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
GOFLAGS=-ldflags="-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"

.PHONY: build clean install test lint

build:
	$(GO) build -o bin/$(BINARY) $(GOFLAGS) ./cmd/dk/

install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/$(BINARY) $(HOME)/.local/bin/

clean:
	rm -rf bin/

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

run: build
	./bin/$(BINARY)
