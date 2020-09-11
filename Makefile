GIT_COMMIT := $(shell git rev-list -1 HEAD)
GO_BUILD := go build -v -ldflags "-X main.GitCommit=$(GIT_COMMIT)"

NAME = check_by_powershell

.PHONY: all clean build test

all: build test

distclean: clean
clean:
	rm -rf build/

build:
	mkdir -p build
	GOOS=linux   GOARCH=amd64 $(GO_BUILD) -o build/$(NAME)-linux-amd64 .
	GOOS=linux   GOARCH=386   $(GO_BUILD) -o build/$(NAME)-linux-i386 .
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o build/$(NAME)-windows-amd64.exe .
	GOOS=darwin  GOARCH=amd64 $(GO_BUILD) -o build/$(NAME)-darwin-amd64 .

test:
	go test -v ./...
