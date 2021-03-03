GOBIN=$(GOPATH)/bin
GOFILES=$(wildcard *.go)
GONAME=$(shell basename "$(PWD)")

build:
	@echo "Building $(GOFILES) to ./bin"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o bin/$(GONAME) .

install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install .

tidy:
	go mod tidy

clean:
	@echo "Cleaning"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean