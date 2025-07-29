SHELL = /bin/zsh

.PHONY: help build install

help:
		@echo "use the makefile with either `build` or `install`"

build:
		@go build -o ftpd ./cmd/client
		@go build -o ftpd-server ./cmd/server

install: build
		@mkdir ~/.ftpd
		@mv ./ftpd ~/.local/bin/
