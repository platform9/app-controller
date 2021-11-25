# Copyright 2021 The pf9ctl authors.
#
# Usage:
# make                 # builds the artifact
# make clean           # removes the artifact and the vendored packages
# make clean-all       # same as make clean + removes the bin dir which houses dep
# make container-build # build artifact on a Linux based container using golang 1.14.1

SHELL := /usr/bin/env bash
BUILD_NUMBER ?= 10
GITHASH := $(shell git rev-parse --short HEAD)
PF9_VERSION ?= 1.0.0
VERSION := $(PF9_VERSION)-$(BUILD_NUMBER)
DETECTED_OS := $(shell uname -s)
BIN_DIR := $(shell pwd)/bin
CMD_DIR := $(shell pwd)/cmd
BIN := fast-path
REPO := fast-path
PACKAGE_GOPATH := /go/src/github.com/platform9/$(REPO)
LDFLAGS := $(shell source ./version.sh ; KUBE_ROOT=. ; KUBE_GIT_VERSION=${VERSION_OVERRIDE} ; kube::version::ldflags)
GIT_STORAGE_MOUNT := $(shell source ./git_utils.sh; container_git_storage_mount)
CONT_USER := $(shell id -u)
CONT_GRP := $(shell id -g)
XDG_CACHE_HOME := /tmp
GOFLAGS ?= ""

.PHONY: clean clean-all default format test

default: $(BIN)

$(BIN): test
	go build -o $(BIN_DIR)/$(BIN) -ldflags "$(LDFLAGS)"

format:
	gofmt -w -s *.go
	gofmt -w -s */*.go

clean:
	rm -rf $(BIN_DIR)

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o $(BIN_DIR)/$(BIN) $(CMD_DIR)/main.go

test:
	go test -v ./...

