.PHONY: fmt test testv cover race build check ci clean

fmt:
	gofmt -w .

test:
	go test ./...

testv:
	go test -v ./...

cover:
	go test -v -cover ./...

race:
	go test -race ./...

build:
	go build ./...

check: fmt testv cover race build

ci: check

clean:
	rm -rf .gocache .gomodcache go-github-actions-demo
