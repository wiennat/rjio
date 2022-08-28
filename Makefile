# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=rjio
FETCH_BINARY_NAME=rjio-fetch

build: build-serve build-fetch

build-serve:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags netgo -a -ldflags '-linkmode external -extldflags "-static"' -o dist/$(BINARY_NAME) -v 

build-fetch:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o dist/$(FETCH_BINARY_NAME) cmd/fetch/main.go

run:
	$(GOBUILD) -o dist/$(BINARY_NAME) main.go
	dist/$(BINARY_NAME)

deps:
	$(GOMOD) download

clean: 
	$(GOCLEAN)
	rm -f ./dist
	
