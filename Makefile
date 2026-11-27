.PHONY: run build fmt vet test clean install service-install service-uninstall \
        widget-test widget-popup widget-install widget-refresh fake

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
	@rm -rf bin

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
