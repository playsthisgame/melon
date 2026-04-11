BINARY   := melon
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: all build install test lint clean

all: build

build:
	go build $(LDFLAGS) -o bin/melon ./cmd/melon
	go build $(LDFLAGS) -o bin/mln   ./cmd/mln

install:
	go install $(LDFLAGS) ./cmd/mln
	go install $(LDFLAGS) ./cmd/melon

test:
	go test ./...

test-verbose:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
