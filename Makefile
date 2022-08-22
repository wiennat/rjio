# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=rjio
BINARY_UNIX=$(BINARY_NAME)_unix

# Rice path
RICE=$(GOPATH)/bin/rice
FETCH_BINARY_NAME=rjio-fetch
FETCH_BINARY_UNIX=$(BINARY_NAME)_unix

build:
	cd feed && $(RICE) embed-go && cd ..
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags netgo -a -ldflags '-linkmode external -extldflags "-static"' -o dist/$(BINARY_NAME) -v 

build-fetch:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o dist/$(FETCH_BINARY_NAME) cmd/fetch/main.go

run:
	# cd feed && $(RICE) embed-go && cd ..
	$(GOBUILD) -o dist/$(BINARY_NAME) main.go
	dist/$(BINARY_NAME)

run-serve:
	# cd feed && $(RICE) embed-go && cd ..
	$(GOBUILD) -o dist/$(BINARY_NAME) main.go
	dist/$(BINARY_NAME) -s true -f false

deps:
	$(GOGET) github.com/GeertJohan/go.rice
	$(GOGET) github.com/GeertJohan/go.rice/rice
	$(GOMOD) download

deps-fetch:
	$(GOMOD) download

clean: 
	$(GOCLEAN)
	rm -f ./dist
	
