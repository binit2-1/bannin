.PHONY: run build fmt vet test clean install service-install service-uninstall \
        widget-test widget-popup widget-install widget-refresh fake tarball rpm release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

run:
	@go run ./cmd/bannin $(ARGS)

build:
	@go build -ldflags "$(LDFLAGS)" -o bin/bannin ./cmd/bannin
	@go build -ldflags "$(LDFLAGS)" -o bin/bannind ./cmd/bannind

fmt:
	@go fmt ./...

vet:
	@go vet ./...

test:
	@go test ./...

clean:
	@rm -rf bin build

install: build
	@install -Dm755 bin/bannin $(HOME)/.local/bin/bannin
	@install -Dm755 bin/bannind $(HOME)/.local/bin/bannind

# Run the service on a made-up memory curve (needs the real service stopped).
fake: build
	@bin/bannind --fake

service-install: install
	@install -Dm644 dist/bannind.service $(HOME)/.config/systemd/user/bannind.service
	@systemctl --user daemon-reload
	@systemctl --user enable bannind.service
	@systemctl --user restart bannind.service
	@systemctl --user --no-pager status bannind.service | head -5

service-uninstall:
	@systemctl --user disable --now bannind.service || true
	@rm -f $(HOME)/.config/systemd/user/bannind.service
	@systemctl --user daemon-reload

widget-test:
	@QT_FORCE_STDERR_LOGGING=1 plasmoidviewer -a ui/plasma -l topedge -f horizontal

widget-install:
	@kpackagetool6 -t Plasma/Applet -i ui/plasma || kpackagetool6 -t Plasma/Applet -u ui/plasma

widget-refresh: widget-install
	@systemctl --user restart plasma-plasmashell

# Popup in a window (planar), with QML errors on the terminal.
widget-popup:
	@QT_FORCE_STDERR_LOGGING=1 plasmoidviewer -a ui/plasma

# ---- Packaging (local only: nothing is uploaded) ----
PKG_VERSION ?= $(shell jq -r .KPlugin.Version ui/plasma/metadata.json)
TARBALL := build/bannin-$(PKG_VERSION).tar.gz

# Source tarball with vendor/ included, so rpmbuild needs no network.
tarball:
	@rm -rf build/src && mkdir -p build/src/bannin-$(PKG_VERSION)
	@git archive HEAD | tar -x -C build/src/bannin-$(PKG_VERSION)
	@cd build/src/bannin-$(PKG_VERSION) && go mod vendor
	@tar -czf $(TARBALL) -C build/src bannin-$(PKG_VERSION)
	@echo $(TARBALL)

# With Go from go.dev instead of Fedora's golang package: make rpm RPMFLAGS=--nodeps
RPMFLAGS ?=
rpm: tarball
	@rpmbuild -ba $(RPMFLAGS) dist/bannin.spec \
		--define "_topdir $(CURDIR)/build/rpm" \
		--define "_sourcedir $(CURDIR)/build" \
		--define "pkg_version $(PKG_VERSION)"
	@ls build/rpm/RPMS/*/*.rpm

# Prebuilt tarballs with install.sh: users need no Go, git or make.
RELEASE_ARCHES ?= x86_64 aarch64
release:
	@rm -rf build/release && mkdir -p build/release
	@sh dist/release.sh $(PKG_VERSION) build/release $(RELEASE_ARCHES)
	@ls build/release
