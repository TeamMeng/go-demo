.PHONY: docker
docker:
	@rm webook || true
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o webook .
	@docker rmi -f TeamMeng/webook:v0.0.1
	@docker build -t TeamMeng/webook:v0.0.2 -f Dockerfile .
