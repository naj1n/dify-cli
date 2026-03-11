_VERSION_RAW := $(shell git describe --tags --always --dirty 2>/dev/null | tr -dc 'a-zA-Z0-9._-')
VERSION ?= $(if $(_VERSION_RAW),$(_VERSION_RAW),dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X dify-cli/cmd.version=$(VERSION) \
	-X dify-cli/cmd.commit=$(COMMIT) \
	-X dify-cli/cmd.date=$(DATE)

.PHONY: build test lint fmt clean install

build:
	go build -ldflags '$(LDFLAGS)' -o dify .

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -l -w .

clean:
	rm -f dify

install:
	go install -ldflags '$(LDFLAGS)' .
