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

## 部署 PostgreSQL

PostgreSQL 在 Kubernetes 中建议使用单独的 `Deployment` 和 `Service`。

- `k8s-postgres-deployment.yaml` 负责启动 PostgreSQL 容器
- `k8s-postgres-service.yaml` 负责在集群内暴露 `5432`
- 数据库 `Service` 建议使用 `ClusterIP`，不要对外暴露为 `LoadBalancer`

`k8s-postgres-deployment.yaml` 至少应包含以下环境变量：

```yaml
env:
  - name: POSTGRES_USER
    value: postgres
  - name: POSTGRES_PASSWORD
    value: postgres
  - name: POSTGRES_DB
    value: todo_db
```

应用容器则需要注入数据库连接串：

```yaml
env:
  - name: PORT
    value: "8080"
  - name: DATABASE_URL
    value: postgres://postgres:postgres@webook-postgresql:5432/todo_db?sslmode=disable
```

部署命令：

```bash
kubectl apply -f k8s-postgres-deployment.yaml
kubectl apply -f k8s-postgres-service.yaml
kubectl rollout status deployment/webook-postgresql
```

## 查看资源

```bash
# 基础资源
kubectl get deployments
kubectl get pods
kubectl get svc
kubectl get endpoints webook
kubectl get endpoints webook-postgresql

# PVC 和 PV（持久卷绑定）
kubectl get pvc                    # 查看 PVC 状态和绑定信息
kubectl get pv                     # 查看 PV 状态
kubectl describe pvc <pvc-name>    # 查看 PVC 详细信息（包括绑定到的 PV）
kubectl describe pv <pv-name>      # 查看 PV 详细信息

# 示例：查看 PostgreSQL PVC 绑定状态
kubectl get pvc -l app=webook-postgresql
kubectl describe pvc webook-postgresql

# 查看 Secret 和 ConfigMap
kubectl get secret
kubectl get configmap

# 查看资源详细事件（排查绑定/创建失败时很有用）
kubectl describe pod <pod-name>
kubectl describe deployment webook-postgresql
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

如果你修改了 Go 代码而不是 YAML，还需要：

1. 重新构建镜像
2. 更新 `k8s-webook-deployment.yaml` 中的镜像 tag
3. 再执行 `kubectl apply`

例如：

```bash
docker build -t teammeng/webook:v0.0.2 .
kubectl apply -f k8s-webook-deployment.yaml
kubectl rollout status deployment/webook
```

如果 Deployment 文件内容没变，`kubectl apply` 会显示 `unchanged`，这时不会触发滚动更新。

## 执行数据库迁移

Kubernetes 中即使 PostgreSQL Pod 已经启动，数据库仍然可能是空的。此时如果直接调用注册接口，可能会看到：

```text
ERROR: relation "users" does not exist (SQLSTATE 42P01)
```

这表示数据库连接已经成功，但迁移脚本还没有执行。

先找到 PostgreSQL Pod：

```bash
kubectl get pods -l app=webook-postgresql
```

然后在项目根目录执行迁移：

```bash
kubectl exec -i <postgres-pod-name> -- psql -U postgres -d todo_db < migrations/20260324131714_initial.sql
```

正常情况下会看到类似输出：

```text
CREATE TABLE
CREATE INDEX
ALTER TABLE
```

执行完成后再测试 `/auth/register`。

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
- PostgreSQL 镜像拉取过程中网络中断，例如 `unexpected EOF`

如果 PostgreSQL Pod 一直报 `ErrImagePull` 或 `ImagePullBackOff`，可以先手动拉镜像：

```bash
docker pull postgres:16
kubectl get pods -w
```

如果 `Deployment` 已经改成 `postgres:16`，但旧 Pod 仍然存在，等待新 Pod 创建并进入 `Running` 即可。

### Exec format error

二进制架构与容器运行时不匹配。例如：在 Apple Silicon 上用 `GOARCH=amd64` 构建，或反之。

修复方式：确认 `GOARCH` 与 `kubectl get nodes` 显示的架构一致后重新构建。

### Connection refused

如果 `kubectl port-forward svc/webook 8080:80` 后访问失败，优先检查：

- Pod 日志是否为 `Listening and serving HTTP on :8080`
- Deployment 是否注入了 `PORT=8080`
- Service 的 `targetPort` 是否为 `8080`

### Invalid credentials

如果 `POST /auth/login` 返回：

```json
{
  "error": "Invalid credentials"
}
```

这通常表示：

- 请求已经成功到达应****用
- 登录逻辑已经执行
- 只是用户名或密码不正确

这类报错通常不是 Kubernetes 端口或 Service 配置问题。

### relation "users" does not exist

如果 `POST /auth/register` 返回：

```json
{
  "error": "Failed to create userERROR: relation \"users\" does not exist (SQLSTATE 42P01)"
}
```

这表示：

- 应用已经成功连接 PostgreSQL
- 但是迁移脚本还没执行
- 数据库中还没有 `users` 表

按上面的“执行数据库迁移”步骤补跑 migration 即可。

## 注意事项

- `kubectl` 只是 Kubernetes 客户端，本身不提供集群
- 当前推荐的本地集群是 Docker Desktop 自带 Kubernetes
- 如果只是本地开发接口，不一定需要 Kubernetes，直接 `docker run` 会更简单
