VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/av ./cmd/av

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/av

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf bin
