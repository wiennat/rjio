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

build:
	cd web/handlers && rice embed-go && cd ../..
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags netgo -a -ldflags '-linkmode external -extldflags "-static"' -o dist/$(BINARY_NAME) -v 

run:
	cd web/handlers && rice embed-go && cd ../..
	$(GOBUILD) -o dist/$(BINARY_NAME) main.go
	dist/$(BINARY_NAME)

deps:
	$(GOGET) github.com/GeertJohan/go.rice
	$(GOGET) github.com/GeertJohan/go.rice/rice
	$(GOMOD) download

clean: 
	$(GOCLEAN)
	rm -f ./dist
	
