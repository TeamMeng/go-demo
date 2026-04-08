.PHONY: docker
docker:
	@rm -f cmd/api/webook
	@GOOS=linux GOARCH=arm64 go build -o cmd/api/webook ./cmd/api
	@docker rmi -f teammeng/webook:v0.0.4 || true
	@docker build -t teammeng/webook:v0.0.5 .
