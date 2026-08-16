BINARY := mindm
DIR    := bin

.PHONY: all build run clean fmt vet test lint install

all: build

build:
	mkdir -p $(DIR)
	CGO_ENABLED=1 go build -o $(DIR)/$(BINARY) ./cmd/mindm

run: build
	sudo $(DIR)/$(BINARY)

clean:
	rm -rf $(DIR)

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

lint:
	staticcheck ./...

install: build
	sudo install -m 755 $(DIR)/$(BINARY) /usr/local/bin/$(BINARY)
