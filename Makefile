.PHONY: fmt test build ci

fmt:
	gofmt -w .

test:
	go test ./...

build:
	go build ./...

ci: test build
