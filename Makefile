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