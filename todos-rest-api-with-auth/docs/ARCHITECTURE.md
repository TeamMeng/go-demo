# Todo REST API 架构文档

## 1. 概述

**Todo REST API with Auth** 是一个基于 Go + Gin + PostgreSQL 构建的待办事项 REST API 服务，用于展示现代化的后端设计实践。项目重点演示了两个差异化特性：**Token 绑定机制**和**滑动窗口限流算法**。

**核心技术栈：**

| 技术 | 用途 |
|------|------|
| Go 1.25+ | 编程语言 |
| Gin | HTTP Web 框架 |
| pgx | PostgreSQL 驱动（参数化查询） |
| JWT (golang-jwt/jwt) | 认证令牌 |
| bcrypt | 密码哈希 |
| Redis + Lua | 滑动窗口限流 |

**设计哲学：简洁、安全、可扩展**。通过分层架构、强制数据隔离和纵深防御策略，确保系统在保持简洁的同时具备良好的安全性和可维护性。

## 2. 架构全景

```
┌──────────────────────────────────────────────────────────────┐
│                          Gin Router                          │
│                 (健康检查 / 认证路由 / 受保护路由)            │
├──────────────────────────────────────────────────────────────┤
│  ┌────────────────┐  ┌────────────────┐  ┌───────────────┐  │
│  │    Handlers    │  │   Middleware   │  │  Repository   │  │
│  │   (传输层)      │  │  (认证/限流)    │  │    (数据层)    │  │
│  │                │  │                │  │               │  │
│  │  TodoHandler   │  │ AuthMiddleware │  │ TodoRepo      │  │
│  │  UserHandler   │  │ RateLimit      │  │ UserRepo      │  │
│  └────────────────┘  └────────────────┘  └───────────────┘  │
├──────────────────────────────────────────────────────────────┤
│                      Models (领域模型)                        │
│                   Todo (int)  │  User (UUID)                 │
├──────────────────────────────────────────────────────────────┤
│                         PostgreSQL                            │
│              (pgx 连接池 + 参数化查询 + CASCADE)             │
└──────────────────────────────────────────────────────────────┘
```

**分层职责：**

- **Handlers（传输层）**：解析 HTTP 请求，调用 Repository 层，返回 JSON 响应
- **Middleware（横切关注点）**：JWT 认证校验、Token 绑定验证、限流控制
- **Repository（数据访问层）**：数据库操作的封装，所有查询强制携带 `user_id` 过滤条件
- **Models（领域模型）**：业务数据结构，通过 JSON tag 控制序列化行为

**请求处理流程：**

```
Client Request
      │
      ▼
┌─────────────────┐
│  Gin Router     │  ─── 健康检查 → 直接响应
└────────┬────────┘
         │ 认证路由 /todos/*
         ▼
┌─────────────────┐
│ AuthMiddleware  │  ─── JWT 解析 → Token 绑定验证 → user_id 注入上下文
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Handler       │  ─── 从上下文获取 user_id → 调用 Repository
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Repository     │  ─── WHERE user_id = $1 强制过滤 → 返回用户数据
└────────┬────────┘
         │
         ▼
    PostgreSQL
```

## 3. 设计亮点

### 3.1 Token 绑定机制

**问题**：传统的 JWT 认证存在 Token 被盗用的风险。攻击者获取 Token 后，可在任意客户端使用。

**方案**：将 JWT Token 与签发时的 `User-Agent` 绑定，验证时对比当前请求的 `User-Agent` 与 Token 中存储的值是否一致。

**实现位置**：`internal/middleware/auth_middleware.go:83-89`

```go
// 登录时将 User-Agent 写入 Claims
claims := jwt.MapClaims{
    "user_id":    user.ID,
    "email":      user.Email,
    "exp":        time.Now().Add(24 * time.Hour).Unix(),
    "user_agent": ctx.Request.UserAgent(),  // ← 绑定到签发时的 UA
}

// 验证时检查 UA 是否匹配
if user_agent, ok := Claims["user_agent"].(string); ok {
    if user_agent != ctx.Request.UserAgent() {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token Claims"})
        ctx.Abort()
        return
    }
}
```

**安全收益**：即使 Token 被盗，攻击者也无法在不同 User-Agent 下使用。这是一种低成本的纵深防御策略。

### 3.2 滑动窗口限流

**问题**：简单的固定窗口计数器（如「每分钟最多 5 次」）在窗口边界存在突发流量问题。例如 1:59 发送 5 个请求后，2:01 又可立即发送 5 个请求，实际 2 秒内处理了 10 个请求。

**方案**：使用 Redis 有序集合（Sorted Set）实现精确的滑动窗口算法。

**Lua 脚本实现**：`internal/ratelimit/slide_window.lua`

```lua
-- 滑动窗口算法核心逻辑
local min = now - window  -- 窗口起始时间

-- 1. 移除窗口外的过期请求
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

-- 2. 统计当前窗口内的请求数量
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

-- 3. 判断是否限流
if cnt >= threshold the********n
    return "true"   -- 限流
else
    redis.call('ZADD', key, now, uid)      -- 记录本次请求
    redis.call('PEXPIRE', key, window)    -- 自动清理过期数据
    return "false"  -- 放行
end
```

**算法图示**：

```
时间轴：◄────────────── 60秒窗口 ──────────────►
        ▲
       now

        ├──旧请求过期──┤
                      │
    ┌──────────────────┼──────────────────┐
    │   窗口内的请求    │   窗口外的请求      │
    │   (计入限流计数)  │   (已清理)         │
    └──────────────────┼──────────────────┘
                       │
                   ZREMRANGEBYSCORE 清除 min 之前的数据

```

**技术优势**：

1. **原子性**：Lua 脚本在 Redis 中单线程执行，避免并发竞争条件
2. **精确性**：滑动窗口而非固定边界，限流精度为毫秒级
3. **内存安全**：`PEXPIRE` 自动清理过期数据，防止 Redis 内存泄漏
4. **可嵌入**：`go:embed` 将 Lua 脚本编译进二进制，部署无额外文件依赖

### 3.3 数据隔离设计

**问题**：多租户场景下，如何确保用户只能访问自己的数据？

**方案**：在 Repository 层强制所有查询携带 `user_id` 过滤条件，Handler 层从 JWT上下文中提取 `user_id` 并注入到所有数据操作中。

**强制过滤示例**：`internal/repository/todo_repository.go`

```go
// GetAllTodos - 即使查询全部，也必须携带 user_id
func GetAllTodos(pool *pgxpool.Pool, userID string) ([]models.Todo, error) {
    query := `
        SELECT id, title, completed, created_at, updated_at, user_id
        FROM todos
        WHERE user_id = $1           -- ← 强制过滤
        ORDER BY created_at DESC
    `
    rows, err := pool.Query(ctx, query, userID)
    // ...
}

// GetToDoByID - 查询单条也必须验证归属
func GetToDoByID(pool *pgxpool.Pool, id int, userID string) (*models.Todo, error) {
    query := `
        SELECT ... FROM todos
        WHERE id = $1 AND user_id = $2   -- ← 双重条件
    `
    // ...
}

// DeleteTodo - 删除也必须验证归属
func DeleteTodo(pool *pgxpool.Pool, id int, userID string) error {
    query := `DELETE FROM todos WHERE id = $1 AND user_id = $2`
    // RowsAffected() == 0 时返回 404
}
```

**信任边界**：Handler 层从 JWT Token 解析出 `user_id`（经过签名验证和过期校验），因此认为是可信的。数据隔离的完整性由 Repository 层保证。

## 4. API 设计

### 4.1 端点概览

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | `/` | 否 | 健康检查 |
| POST | `/auth/register` | 否 | 用户注册 |
| POST | `/auth/login` | 否 | 登录（可限流） |
| POST | `/todos` | JWT | 创建待办 |
| GET | `/todos` | JWT | 获取全部 |
| GET | `/todos/:id` | JWT | 获取单条 |
| PUT | `/todos/:id` | JWT | 更新 |
| DELETE | `/todos/:id` | JWT | 删除 |

### 4.2 认证流程时序图

```
┌────┐          ┌────────┐         ┌────────┐         ┌──────────┐
│Client│         │ Router │         │ Handler│         │   DB     │
└──┬───┘         └───┬────┘         └───┬────┘         └────┬────┘
   │                 │                   │                   │
   │ POST /auth/login│                   │                   │
   │───────────────►│                   │                   │
   │                 │                   │                   │
   │                 │ GetUserByEmail    │                   │
   │                 │─────────────────►│                   │
   │                 │                 │ SELECT ...         │
   │                 │                 │───────────────────►│
   │                 │                 │                   │
   │                 │                 │◄───────────────────│
   │                 │◄───────────────│ user record        │
   │                 │                   │                   │
   │                 │ bcrypt.Compare   │                   │
   │                 │──────────────────│                   │
   │                 │                   │                   │
   │  JWT + UA绑定    │                   │                   │
   │◄───────────────│                   │                   │
   │                 │                   │                   │
```

### 4.3 错误处理规范

所有错误响应遵循统一格式：

```json
{
    "error": "Human-readable error message"
}
```

**HTTP 状态码使用规范**：

| 状态码 | 场景 |
|--------|------|
| 200 | 操作成功 |
| 201 | 资源创建成功 |
| 400 | 请求参数错误、邮箱已注册、密码太短 |
| 401 | 未提供 Token、Token 无效/过期/UA 不匹配 |
| 404 | 资源不存在 |
| 429 | 请求被限流 |
| 500 | 服务器内部错误 |

**安全设计**：登录失败时统一返回「Invalid credentials」，避免通过不同错误消息进行用户枚举攻击。

## 5. 安全模型

### 5.1 三层防护体系

```
┌─────────────────────────────────────────────────────┐
│                    第一层：认证                       │
│  JWT 签名验证 + 过期时间校验 + Algorithm 强制校验   │
│  (防止 alg:none、RS256 降级攻击)                    │
├─────────────────────────────────────────────────────┤
│                    第二层：授权                       │
│  Token 绑定 User-Agent + user_id 数据隔离           │
│  (防止 Token 被盗用、越权访问)                      │
├─────────────────────────────────────────────────────┤
│                    第三层：限流                       │
│  Redis 滑动窗口限流 + 统一错误消息                   │
│  (防止暴力破解、DoS 攻击)                          │
└─────────────────────────────────────────────────────┘
```

### 5.2 密码安全

使用 bcrypt 进行密码哈希（成本因子默认 10），密码不以明文存储或传输。

```go
// 存储时哈希
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

// 验证时比对
bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
```

### 5.3 SQL 注入防护

所有数据库操作使用 pgx 的参数化查询：

```go
// 正确：参数化查询
pool.QueryRow(ctx, "SELECT ... WHERE email = $1", email)

// 错误：字符串拼接（存在 SQL 注入风险）
pool.QueryRow(ctx, "SELECT ... WHERE email = '"+email+"'")
```

### 5.4 JWT 安全

1. **算法强制校验**：仅允许 `HS256`，拒绝其他算法（防止算法降级攻击）
2. **过期时间校验**：显式检查 `exp` 字段
3. **Token 绑定**：`user_agent` 存储在 Claims 中，验证时对比

## 6. 数据库架构

### 6.1 ER 关系图

```
┌─────────────────┐         ┌─────────────────┐
│     users       │         │      todos      │
├─────────────────┤         ├─────────────────┤
│ id (PK, UUID)   │◄───────┐│ id (PK, SERIAL) │
│ email (UNIQUE)  │   1:N  ││ title           │
│ password        │        ││ completed       │
│ created_at      │        ││ user_id (FK)────┘
│ updated_at      │        ││ created_at      │
└─────────────────┘        ││ updated_at      │
                           └─────────────────┘
```

### 6.2 表结构说明

**users 表**：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | UUID | PK, DEFAULT gen_random_uuid() | 主键 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | 登录账号 |
| password | VARCHAR(255) | NOT NULL | bcrypt 哈希 |
| created_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**todos 表**：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| title | VARCHAR(255) | NOT NULL | 待办标题 |
| completed | BOOLEAN | DEFAULT FALSE | 完成状态 |
| user_id | UUID | FK → users(id), ON DELETE CASCADE | 所属用户 |
| created_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

### 6.3 索引策略

```sql
-- users 表唯一索引，加速登录查询
CREATE INDEX idx_users_email ON users(email);
```

### 6.4 级联删除

```sql
ALTER TABLE todos ADD CONSTRAINT fk_todos_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
```

删除用户时，该用户的所有待办事项自动被清理，无需手动处理孤儿数据。

## 7. 部署架构

### 7.1 单容器部署

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api ./cmd/api

FROM alpine:latest
COPY --from=builder /app/api /api
COPY --from=builder /app/.env .env
CMD ["/api"]
```

**环境变量要求**：

| 变量 | 说明 |
|------|------|
| DATABASE_URL | PostgreSQL 连接字符串 |
| PORT | HTTP 服务端口（默认 8080） |
| JWT_SECRET | JWT 签名密钥（生产环境应使用强随机字符串） |

### 7.2 Docker Compose 本地开发

```yaml
services:
  api:
    build: .
    ports:
      - "8000:8000"
    env_file:
      - .env
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: todo_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### 7.3 Kubernetes 部署架构

本项目使用 Kubernetes（K8s）进行容器编排，所有资源定义在项目根目录的 `k8s-*.yaml` 文件中。

#### 7.3.1 资源一览表

| 文件 | 资源类型 | 资源名称 | 用途 | 关联资源 |
|------|---------|---------|------|---------|
| `k8s-ingress-nginx.yaml` | Ingress | `webook-ingress` | HTTP/HTTPS 入口路由 | → `webook` Service |
| `k8s-webook-deployment.yaml` | Deployment | `webook` | Webook 应用（3 副本） | ← `webook` Service |
| `k8s-webook-service.yaml` | Service | `webook` | 应用访问入口（LoadBalancer） | ← Ingress, → webook Pods |
| `k8s-postgres-deployment.yaml` | Deployment | `webook-postgresql` | PostgreSQL 数据库 | ← `postgres-claim` PVC |
| `k8s-postgres-service.yaml` | Service | `webook-postgresql` | 数据库 DNS 访问（ClusterIP） | → PostgreSQL Pods |
| `k8s-postgres-pv.yaml` | PersistentVolume | `my-local-pv` | 本地持久化存储 | → `postgres-claim` PVC |
| `k8s-postgres-pvc.yaml` | PersistentVolumeClaim | `postgres-claim` | 存储声明 | ← PostgreSQL Deployment |
| `k8s-redis-deployment.yaml` | Deployment | `webook-redis` | Redis 缓存（单副本） | ← `webook-redis` Service |
| `k8s-redis-service.yaml` | Service | `webook-redis` | 缓存访问入口（ClusterIP） | → Redis Pods |

#### 7.3.2 资源依赖关系图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Ingress (webook-ingress)               │
│                     host: live.webook.com                      │
│                     k8s-ingress-nginx.yaml                     │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                │ 路由到
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Service (webook)                             │
│                    k8s-webook-service.yaml                      │
│                    type: LoadBalancer | port: 80 → 8080        │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                    ┌───────────┼───────────┐
                    │           │           │
                    ▼           ▼           ▼
            ┌───────────┐ ┌───────────┐ ┌───────────┐
            │ Deployment│ │ Deployment│ │ Deployment│
            │  webook   │ │  webook   │ │  webook   │
            │ (replica) │ │ (replica) │ │ (replica) │
            └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
                  │           │           │
                  └───────────┼───────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │  Service        │ │  Service        │ │  Env Variables  │
    │  webook-postgres│ │  webook-redis   │ │  DATABASE_URL   │
    │  k8s-postgres-  │ │  k8s-redis-     │ │  REDIS_URL      │
    │  service.yaml   │ │  service.yaml   │ │                 │
    └────────┬────────┘ └────────┬────────┘ └─────────────────┘
             │                   │
             ▼                   ▼
    ┌─────────────────┐ ┌─────────────────┐
    │   Deployment     │ │   Deployment    │
    │ webook-postgres  │ │  webook-redis  │
    │ k8s-postgres-    │ │ k8s-redis-      │
    │ deployment.yaml  │ │ deployment.yaml │
    └────────┬────────┘ └─────────────────┘
             │
             ▼
    ┌─────────────────┐
    │      PVC        │
    │  postgres-claim │
    │ k8s-postgres-   │
    │ pvc.yaml        │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │       PV        │
    │   my-local-pv   │
    │ k8s-postgres-   │
    │ pv.yaml         │
    └─────────────────┘
```

#### 7.3.3 部署顺序

```bash
# 1. 部署 PostgreSQL（先启动数据库）
kubectl apply -f k8s-postgres-pv.yaml
kubectl apply -f k8s-postgres-pvc.yaml
kubectl apply -f k8s-postgres-deployment.yaml
kubectl apply -f k8s-postgres-service.yaml

# 2. 部署 Redis
kubectl apply -f k8s-redis-deployment.yaml
kubectl apply -f k8s-redis-service.yaml

# 3. 部署 Webook 应用
kubectl apply -f k8s-webook-deployment.yaml
kubectl apply -f k8s-webook-service.yaml

# 4. 部署 Ingress（可选，需要提前安装 Nginx Ingress Controller）
kubectl apply -f k8s-ingress-nginx.yaml

# 5. 验证部署状态
kubectl get pods -l app=webook
kubectl get svc
kubectl get pvc
```

#### 7.3.4 关键配置说明

**Webook Deployment 环境变量：**

```yaml
env:
  - name: PORT
    value: "8080"                                    # 应用监听端口
  - name: DATABASE_URL
    value: postgres://postgres:postgres@webook-postgresql:5432/todo_db?sslmode=disable
                                                          # PostgreSQL 连接地址（使用 K8s Service DNS）
  - name: REDIS_URL
    value: webook-redis:6379                         # Redis 连接地址
```

**PostgreSQL Deployment 存储配置：**

```yaml
volumeMounts:
  - mountPath: /var/lib/postgresql/data
    name: postgres-storage
volumes:
  - name: postgres-storage
    persistentVolumeClaim:
      claimName: postgres-claim    # 关联 k8s-postgres-pvc.yaml
```

#### 7.3.5 访问方式

| 环境 | 访问方式 |
|------|---------|
| Docker Desktop K8s | `kubectl port-forward svc/webook 8080:80` → `http://127.0.0.1:8080` |
| 有 LoadBalancer 的云环境 | `http://<LB-IP>/` |
| 配置 Ingress 后 | `http://live.webook.com/`（需配置 hosts） |

## 8. 扩展性与未来方向

### 8.1 当前局限性

1. **会话管理**：Token 签发后无法主动撤销，只能等待过期
2. **限流粒度**：默认基于 IP，可扩展为基于用户 ID
3. **没有刷新 Token**：24 小时过期后需重新登录
4. **没有权限管理**：所有认证用户具有同等权限

### 8.2 可选扩展路径

**短期扩展**：
- 添加 Refresh Token 实现无感续期
- 引入 Redis 黑名单实现主动 Token 撤销
- 增加 Rate Limit 配置的动态更新（无需重启）

**中期扩展**：
- 添加 OAuth2 第三方登录（GitHub、Google）
- 实现 WebSocket 实时通知
- 引入缓存层（Redis Cache）优化读性能

**长期扩展**：
- 微服务拆分（Auth Service / Todo Service）
- 引入消息队列处理异步任务
- 多租户支持（通过 Schema 隔离）

## 9. 项目结构参考

```
todos-rest-api-with-auth/
├── cmd/api/
│   └── main.go                 # 应用入口，路由注册
├── internal/
│   ├── config/                 # 配置加载
│   │   └── config.go
│   ├── database/               # PostgreSQL 连接池
│   │   └── postgres.go
│   ├── handlers/                # HTTP 处理器
│   │   ├── todo_handler.go     # Todo CRUD
│   │   └── user_handler.go     # 用户注册/登录
│   ├── middleware/             # 中间件
│   │   └── auth_middleware.go  # JWT 认证 + Token 绑定
│   ├── ratelimit/              # 限流核心
│   │   ├── types.go            # Limiter 接口
│   │   ├── redis_slide_window.go
│   │   └── slide_window.lua    # Lua 脚本（嵌入二进制）
│   ├── models/                 # 领域模型
│   │   ├── todo.go
│   │   └── user.go
│   └── repository/             # 数据访问层
│       ├── todo_repository.go
│       └── user_repository.go
├── migrations/                 # 数据库迁移
│   └── 20260324131714_initial.sql
├── docs/                       # 文档
│   ├── ARCHITECTURE.md         # 架构文档
│   ├── docker-k8s-commands.md   # Docker/K8s 命令速查
│   └── redis-ip-ratelimit.md   # Redis 限流实现
├── k8s-*.yaml                  # Kubernetes 资源定义
│   ├── k8s-ingress-nginx.yaml     # Ingress（外部访问路由）
│   ├── k8s-webook-deployment.yaml  # Webook Deployment
│   ├── k8s-webook-service.yaml     # Webook Service
│   ├── k8s-postgres-deployment.yaml# PostgreSQL Deployment
│   ├── k8s-postgres-service.yaml   # PostgreSQL Service
│   ├── k8s-postgres-pv.yaml       # PostgreSQL PersistentVolume
│   ├── k8s-postgres-pvc.yaml      # PostgreSQL PersistentVolumeClaim
│   ├── k8s-redis-deployment.yaml  # Redis Deployment
│   └── k8s-redis-service.yaml     # Redis Service
├── .env                        # 环境变量
└── go.mod
```
