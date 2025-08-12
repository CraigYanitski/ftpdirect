.PHONY: build build-exec install

help:
		@echo "FTPdirect (aka ftpd) is a peer-to-peer file transfer program that is currently under development. "
		@echo "The idea is to enable port forwarding for peers behind NATs (which should be every peer) in order "
		@echo "to permit direct TCP connections."
		@echo ""
		@echo "Make the project with either the build or install phony targets depending on whether you are "
		@echo "testing the program from this repo or using it as it is intended."
		@echo ""
		@echo "The phony target build-exec is useful to build executables if you are developing your own instance."
		@echo ""
		@echo "!!  Unfortunately installation currently only works on linux and macOS  !!"
		@echo ""

build:
		@go install github.com/go-gost/gost/cmd/gost@latest
		@go build -o ftpd ./cmd/client
		@go build -o ftpd-server ./cmd/server

build-exec:
		@GOOS=linux GOARCH=amd64 go build -o ftpd ./cmd/client
		@GOOS=windows GOARCH=amd64 go build -o ftpd-win.exe ./cmd/client
		@GOOS=darwin GOARCH=amd64 go build -o ftpd-mac ./cmd/client
		@go build -o ftpd-server ./cmd/server

install: build
		@mkdir -p ~/.ftpd
		@mv ./ftpd ~/.local/bin/
