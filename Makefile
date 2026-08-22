BINARY := minidm
BIN_DIR := bin

.PHONY: build run clean fmt fmt-check vet test staticcheck lint install

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/minidm

run: build
	sudo ./$(BIN_DIR)/$(BINARY)

fmt:
	go fmt ./...

fmt-check:
	test -z "$$(gofmt -s -l .)"

vet:
	go vet ./...

test:
	go test ./...

staticcheck:
	staticcheck ./...

lint: fmt-check vet staticcheck

install:
	@set -e; \
	TMPDIR=$$(mktemp -d); \
	cp package/PKGBUILD "$$TMPDIR/"; \
	cd "$$TMPDIR"; \
	makepkg -si; \
	STATUS=$$?; \
	cd "$$OLDPWD"; \
	rm -rf "$$TMPDIR"; \
	exit $$STATUS

clean:
	rm -rf $(BIN_DIR)
