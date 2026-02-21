.PHONY: all build test lint clean generate install docker release

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X spawn.dev/internal/buildinfo.Version=$(VERSION)"

all: lint test build

build:
	go build $(LDFLAGS) -o bin/spawn ./cmd/spawn
	go build $(LDFLAGS) -o bin/spawnd ./cmd/spawnd
	go build $(LDFLAGS) -o bin/spawn-sandbox ./cmd/spawn-sandbox

test:
	go test -cover ./...

test-integration:
	go test -tags=integration ./test/...

lint:
	go test ./...

generate:
	go generate ./...

clean:
	rm -rf bin/

install:
	go install $(LDFLAGS) ./cmd/spawn

docker:
	docker build -t spawn:$(VERSION) .

release:
	goreleaser release --clean
