# 基础镜像
FROM ubuntu:20.04

# 把编译后的打包进来这个镜像
COPY cmd/api/webook /app/webook
WORKDIR /app

ENTRYPOINT ["/app/webook"]
