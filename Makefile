SHELL = /bin/sh

.PHONY: help build build-exec install

help:
		@echo "use the makefile with either `build` or `install`"

build:
		@go build -o ftpd ./cmd/client
		@go build -o ftpd-server ./cmd/server

build-exec:
		@GOOS=linux GOARCH=amd64 go build -o ftpd ./cmd/client
		@GOOS=windows GOARCH=amd64 go build -o ftpd-win ./cmd/client
		@GOOS=darwin GOARCH=amd64 go build -o ftpd-mac ./cmd/client
		@go build -o ftpd-server ./cmd/server

install: build
		@mkdir -p ~/.ftpd
		@mv ./ftpd ~/.local/bin/
