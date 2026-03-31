# Todo REST API with Auth

一个基于 Go + Gin + PostgreSQL 构建的待办事项 REST API 服务，支持用户注册、JWT 认证和完整的 CRUD 操作。

## 功能特性

- **用户认证**：基于 JWT 的认证机制，支持用户注册和登录
- **数据隔离**：每个用户只能访问自己的待办事项
- **RESTful API**：标准的 REST 接口设计
- **数据库支持**：PostgreSQL 持久化存储
- **密码安全**：使用 bcrypt 进行密码哈希

## 技术栈

| 技术 | 用途 |
|------|------|
| [Go](https://go.dev/) | 编程语言 |
| [Gin](https://gin-gonic.com/) | HTTP Web 框架 |
| [pgx](https://github.com/jackc/pgx) | PostgreSQL 驱动 |
| [JWT](https://github.com/golang-jwt/jwt) | 认证令牌 |
| [bcrypt](https://golang.org/x/crypto/bcrypt) | 密码哈希 |
| [godotenv](https://github.com/joho/godotenv) | 环境变量管理 |

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
│   │   └── auth_middleware.go  # JWT 认证中间件
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

## 许可证

MIT License
