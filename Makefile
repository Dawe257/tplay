PREFIX ?= $(HOME)/.local
VERSION ?= $(patsubst v%,%,$(shell git describe --tags --always --dirty 2>/dev/null || echo dev))
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build test lint install clean

build:
	go build -ldflags "$(LDFLAGS)" -o tplay ./cmd/tplay

test:
	go test -race ./...

lint:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	go vet ./...

install:
	go build -ldflags "$(LDFLAGS)" -o $(PREFIX)/bin/tplay ./cmd/tplay

clean:
	rm -f tplay
