.PHONY: run build fmt vet

run:
	@go run ./cmd/bannin $(ARGS)

build:
	@go build -o bin/bannin ./cmd/bannin

fmt:
	@go fmt ./...

vet:
	@go vet ./...

clean:
	@rm -rf bin

install: build
	@install -Dm755 bin/bannin $(HOME)/.local/bin/bannin

widget-test:
	@plasmoidviewer -a ui/plasma -l topedge -f horizontal

widget-install:
	@kpackagetool6 -t Plasma/Applet -i ui/plasma || kpackagetool6 -t Plasma/Applet -u ui/plasma

widget-refresh: widget-install
	@systemctl --user restart plasma-plasmashell