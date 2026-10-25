PREFIX ?= $(HOME)/.local

.PHONY: build test lint install clean

build:
	go build -o tplay ./cmd/tplay

test:
	go test -race ./...

lint:
	gofmt -l .
	go vet ./...

install:
	go build -o $(PREFIX)/bin/tplay ./cmd/tplay

clean:
	rm -f tplay
