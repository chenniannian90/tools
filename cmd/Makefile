PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = v$(shell cat version/version)
COMMIT_SHA ?= $(shell git describe --always)-devel

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOBUILD=CGO_ENABLED=0 go build -ldflags "-X ${PKG}/version.Version=${VERSION}+sha.${COMMIT_SHA}"
GOINSTALL=CGO_ENABLED=0 go install -ldflags "-X ${PKG}/version.Version=${VERSION}+sha.${COMMIT_SHA}"

MAIN_ROOT ?= ./cmd/tools

install: download
	cd $(MAIN_ROOT) && $(GOINSTALL)

build:
	cd $(MAIN_ROOT) && $(GOBUILD) -o tools

download:
	go mod download -x

dep:
	go get -u ./...
