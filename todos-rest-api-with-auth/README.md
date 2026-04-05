# Todo REST API with Auth

一个基于 Go + Gin + PostgreSQL 构建的待办事项 REST API 服务，支持用户注册、JWT 认证和完整的 CRUD 操作。

## 功能特性

- **用户认证**：基于 JWT 的认证机制，支持用户注册和登录
- **数据隔离**：每个用户只能访问自己的待办事项
- **RESTful API**：标准的 REST 接口设计
- **数据库支持**：PostgreSQL 持久化存储
- **密码安全**：使用 bcrypt 进行密码哈希
- **请求限流**：基于 Redis 滑动窗口算法的限流机制，保护 API 免受过度请求

## 技术栈

| 技术 | 用途 |
|------|------|
| [Go](https://go.dev/) | 编程语言 |
| [Gin](https://gin-gonic.com/) | HTTP Web 框架 |
| [pgx](https://github.com/jackc/pgx) | PostgreSQL 驱动 |
| [JWT](https://github.com/golang-jwt/jwt) | 认证令牌 |
| [bcrypt](https://golang.org/x/crypto/bcrypt) | 密码哈希 |
| [godotenv](https://github.com/joho/godotenv) | 环境变量管理 |
| [Redis](https://redis.io/) | 限流计数器存储 |
| [go-redis](https://github.com/redis/go-redis) | Redis Go 客户端 |

## 项目结构

```
todos-rest-api-with-auth/
├── cmd/api/
│   └── main.go                 # 应用入口
├── internal/
│   ├── config/                 # 配置管理
│   │   └── config.go
│   ├── database/               # 数据库连接
│   │   └── postgres.go
│   ├── handlers/               # HTTP 请求处理器
│   │   ├── todo_handler.go     # Todo CRUD 接口
│   │   └── user_handler.go     # 用户认证接口
│   ├── middleware/             # 中间件
│   │   ├── auth_middleware.go  # JWT 认证中间件
│   │   └── ratelimit/          # 限流中间件
│   │       └── builder.go       # 限流中间件构建器
│   ├── ratelimit/              # 限流核心实现
│   │   ├── types.go            # 限流器接口定义
│   │   ├── redis_slide_window.go # Redis 滑动窗口限流器
│   │   └── slide_window.lua    # Lua 脚本（嵌入到 Go）
│   ├── models/                 # 数据模型
│   │   ├── todo.go
│   │   └── user.go
│   └── repository/             # 数据访问层
│       ├── todo_repository.go
│       └── user_repository.go
├── migrations/                 # 数据库迁移脚本
│   └── 20260324131714_initial.sql
├── .env                        # 环境变量配置
├── go.mod
└── README.md
```

## 环境要求

- Go 1.25+
- PostgreSQL 14+
- Redis 6+
- (可选) make

## 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd todos-rest-api-with-auth
```

### 2. 创建数据库

```bash
# 使用 psql 创建数据库
psql -U postgres -c "CREATE DATABASE todo_db;"

# 或使用 createdb 命令
createdb todo_db
```

### 3. 配置环境变量

复制 `.env` 文件并根据实际情况修改：

```bash
# 数据库连接字符串
DATABASE_URL="postgres://postgres:postgres@localhost:5432/todo_db"

# HTTP 服务端口
PORT=8000

# JWT 签名密钥（生产环境请使用强随机字符串）
JWT_SECRET="my-jwt-password-string-security"
```

### 4. 执行数据库迁移

```bash
# 使用 psql 执行迁移脚本
psql -U postgres -d todo_db -f migrations/20260324131714_initial.sql
```

### 5. 安装依赖并运行

```bash
# 下载依赖
go mod download

# 运行服务
go run cmd/api/main.go
```

服务启动后，访问 http://localhost:8000/ 看到欢迎消息即表示成功。

### 6. 配置 Redis（限流功能）

限流功能依赖 Redis，需确保 Redis 服务正在运行：

```bash
# macOS (使用 Homebrew)
brew services start redis

# 或 Docker 运行
docker run -d -p 6379:6379 redis:latest
```

**限流配置说明：**

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| 窗口时间 | 滑动窗口的时间范围 | 1 分钟 |
| 请求限额 | 窗口内允许的最大请求数 | 5 次 |

**在代码中使用限流中间件：**

```go
// 创建 Redis 客户端
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建滑动窗口限流器（1分钟内最多5次请求）
limiter := ratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 5)

// 使用构建器创建中间件
rateMiddleware := ratelimit.NewBuilder(limiter).Build()

// 注册到路由
router.POST("/auth/login", rateMiddleware, handlers.LoginHandler(pool, cfg))
```

**自定义限流 Key：**

默认使用客户端 IP 作为限流 key，可自定义：

```go
// 按用户ID限流
rateMiddleware := ratelimit.NewBuilder(limiter).
    SetKeyGenFunc(func(ctx *gin.Context) string {
        userID := ctx.GetHeader("X-User-ID")
        return "user-limit:" + userID
    }).
    Build()
```

## API 文档

### 公开端点（无需认证）

#### 用户注册

```http
POST /auth/register
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

**响应：**
```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
}
```

#### 用户登录

```http
POST /auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "password123"
}
```

**响应：**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**限流响应（429 Too Many Requests）：**

当请求超过限流阈值时，返回以下响应：

```json
{
    "message": "Too Many Requests"
}
```

**限流说明：**

- 登录接口默认限流：1 分钟内最多 5 次请求
- 使用 Redis 滑动窗口算法，确保高精度限流
- 限流 key 默认基于客户端 IP 地址

### 受保护端点（需要 JWT Token）

所有 `/todos` 下的端点都需要在请求头中携带 JWT Token：

```http
Authorization: Bearer <your-jwt-token>
```

#### 创建待办事项

```http
POST /todos
Content-Type: application/json
Authorization: Bearer <token>

{
    "title": "Buy groceries",
    "completed": false
}
```

**响应：**
```json
{
    "id": 1,
    "title": "Buy groceries",
    "completed": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### 获取所有待办事项

```http
GET /todos
Authorization: Bearer <token>
```

**响应：**
```json
[
    {
        "id": 1,
        "title": "Buy groceries",
        "completed": false,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "user_id": "550e8400-e29b-41d4-a716-446655440000"
    }
]
```

#### 获取单个待办事项

```http
GET /todos/1
Authorization: Bearer <token>
```

#### 更新待办事项

```http
PUT /todos/1
Content-Type: application/json
Authorization: Bearer <token>

{
    "title": "Buy groceries and cook dinner",
    "completed": true
}
```

**注意：** `title` 和 `completed` 字段都是可选的，可以只更新其中一个。

#### 删除待办事项

```http
DELETE /todos/1
Authorization: Bearer <token>
```

**响应：**
```json
{
    "message": "Todo deleted successfully"
}
```

### 健康检查

```http
GET /
```

**响应：**
```json
{
    "message": "Todo API is running",
    "status": "success",
    "database": "connected"
}
```

## 数据库设计

### users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键，自动生成 |
| email | VARCHAR(255) | 唯一，用于登录 |
| password | VARCHAR(255) | bcrypt 哈希后的密码 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

### todos 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 自增主键 |
| title | VARCHAR(255) | 待办事项标题 |
| completed | BOOLEAN | 完成状态 |
| user_id | UUID | 外键，关联 users 表 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

## 开发说明

### 运行测试

```bash
go test -v ./...
```

### 代码格式化

```bash
gofmt -w .
```

### 构建可执行文件

```bash
# Linux/Mac
go build -o todo-api cmd/api/main.go

# Windows
go build -o todo-api.exe cmd/api/main.go
```

## 常见问题

### Q: 启动时报 "Failed to connect to database" 错误？

A: 请检查：
1. PostgreSQL 服务是否已启动
2. `.env` 中的 `DATABASE_URL` 配置是否正确
3. 数据库 `todo_db` 是否已创建

### Q: 登录时返回 "Invalid or expired token"？

A: 可能原因：
1. JWT Token 已过期（默认 24 小时）
2. 服务端 `JWT_SECRET` 与签发时不一致
3. Token 格式不正确（确保使用 `Bearer <token>` 格式）

### Q: 注册时提示 "Email already registered"？

A: 邮箱地址具有唯一约束，该邮箱已被其他用户注册。

### Q: 登录时返回 429 Too Many Requests？

A: 这是限流保护机制，说明在时间窗口内请求次数超过了限制。默认配置为每分钟最多 5 次登录尝试。等待一分钟后重试即可。

### Q: 限流功能不生效？

A: 请检查：
1. Redis 服务是否正在运行
2. Redis 连接地址是否正确（默认 localhost:6379）
3. 检查 Redis 是否可以正常连接

## 许可证

MIT License
