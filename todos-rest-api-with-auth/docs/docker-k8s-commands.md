# Docker Desktop Kubernetes 常用命令

## 适用场景

本文件适用于当前项目在本地使用以下环境开发和验证：

- Docker Desktop
- Docker Desktop 自带 Kubernetes
- `kubectl`

如果你只是想本地跑 HTTP 服务，不一定需要 Kubernetes，可以直接使用 `docker run`。

## 前置检查

确认 Docker CLI 正在使用 Docker Desktop：

```bash
docker context use desktop-linux
docker version
kubectl config current-context
kubectl get nodes
```

正常情况下：

- `docker version` 的服务端应显示 `Docker Desktop`
- `kubectl get nodes` 应返回本地 Kubernetes 节点

## 部署应用

```bash
kubectl apply -f k8s-webook-deployment.yaml
kubectl apply -f k8s-webook-service.yaml
kubectl rollout status deployment/webook
```

## 查看资源

```bash
kubectl get deployments
kubectl get pods
kubectl get svc
kubectl get endpoints webook
```

## 删除应用

```bash
kubectl delete -f k8s-webook-service.yaml
kubectl delete -f k8s-webook-deployment.yaml
```

## 构建镜像并部署流程

### 1. 确认架构

```bash
# Apple Silicon 为 arm64，Intel 为 amd64
uname -m

# 查看 Kubernetes 节点架构
kubectl get nodes -o wide
```

Go 二进制必须与容器运行时架构匹配，否则会报 `exec format error`。

| 目标平台 | GOARCH |
|---------|--------|
| Apple Silicon Mac / Docker Desktop Kubernetes | `arm64` |
| Intel Mac / Docker Desktop Kubernetes | `amd64` |
| 树莓派 (32位) | `arm` |

### 2. 构建 Go 二进制

```bash
# Apple Silicon
GOOS=linux GOARCH=arm64 go build -o cmd/api/webook ./cmd/api

# Intel
GOOS=linux GOARCH=amd64 go build -o cmd/api/webook ./cmd/api
```

### 3. 构建 Docker 镜像

```bash
docker build -t teammeng/webook:v0.0.1 .
```

### 4. 部署到 Docker Desktop Kubernetes

Docker Desktop 自带 Kubernetes 通常可以直接使用本地 Docker 镜像，不需要再执行 `minikube image load` 这类步骤。

```bash
kubectl apply -f k8s-webook-deployment.yaml
kubectl apply -f k8s-webook-service.yaml
kubectl rollout status deployment/webook
```

### 5. 验证

```bash
kubectl get pods -l app=webook
kubectl logs -l app=webook --tail=20
```

正常情况下日志里应看到：

```text
Listening and serving HTTP on :8080
```

## 访问服务

在 Docker Desktop Kubernetes 下，如果本机已经可以直接访问 `LoadBalancer` 暴露的端口，可以直接验证：

```bash
curl -i http://localhost:80/
```

如果你的本机环境下不能直接访问 `localhost:80`，开发阶段最稳的方式是：

```bash
kubectl port-forward svc/webook 8080:80
```

然后在另一个终端访问：

```bash
curl -i http://127.0.0.1:8080/
```

## 不用 Kubernetes 的本地运行方式

如果你只是要本地开发接口，可以直接运行容器：

```bash
docker build -t teammeng/webook:v0.0.1 .
docker run --rm -p 8080:8080 -e PORT=8080 teammeng/webook:v0.0.1
```

然后访问：

```bash
curl -i http://127.0.0.1:8080/
```

## 常见问题

### ImagePullBackOff

Kubernetes 找不到镜像。常见原因：

- 镜像未构建
- 镜像 tag 与 Deployment 中写的不一致
- 当前 Kubernetes 集群不是 Docker Desktop，自然也看不到本地 Docker 镜像

### Exec format error

二进制架构与容器运行时不匹配。例如：在 Apple Silicon 上用 `GOARCH=amd64` 构建，或反之。

修复方式：确认 `GOARCH` 与 `kubectl get nodes` 显示的架构一致后重新构建。

### Connection refused

如果 `kubectl port-forward svc/webook 8080:80` 后访问失败，优先检查：

- Pod 日志是否为 `Listening and serving HTTP on :8080`
- Deployment 是否注入了 `PORT=8080`
- Service 的 `targetPort` 是否为 `8080`

## 注意事项

- `kubectl` 只是 Kubernetes 客户端，本身不提供集群
- 当前推荐的本地集群是 Docker Desktop 自带 Kubernetes
- 如果只是本地开发接口，不一定需要 Kubernetes，直接 `docker run` 会更简单
