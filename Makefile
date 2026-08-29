BINARY := minidm
BIN_DIR := bin
PREFIX ?= /usr

.PHONY: all build run fmt fmt-check vet test staticcheck lint install package clean

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -trimpath -buildvcs=false -o $(BIN_DIR)/$(BINARY) ./cmd/minidm

run: build
	sudo ./$(BIN_DIR)/$(BINARY)

fmt:
	gofmt -s -w .

fmt-check:
	test -z "$$(gofmt -s -l .)"

vet:
	go vet ./...

test:
	go test ./...

staticcheck:
	staticcheck ./...

lint: fmt-check vet staticcheck

install: build
	install -Dm755 $(BIN_DIR)/$(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	install -Dm644 package/$(BINARY).service $(DESTDIR)/usr/lib/systemd/system/$(BINARY).service
	install -Dm644 package/pam.d/minidm $(DESTDIR)/etc/pam.d/minidm
	install -Dm644 LICENSE $(DESTDIR)$(PREFIX)/share/licenses/$(BINARY)/LICENSE

package:
	@tmpdir=$$(mktemp -d); \
	cp package/PKGBUILD "$$tmpdir/"; \
	(cd "$$tmpdir" && makepkg -si); \
	status=$$?; \
	rm -rf "$$tmpdir"; \
	exit $$status

clean:
	rm -rf $(BIN_DIR)
