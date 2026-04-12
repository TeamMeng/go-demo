// package models 定义应用的核心数据结构（领域模型）。
//
// 该包中的结构体被仓库层、处理器层及序列化层共用，
// 通过 json tag 控制对外 API 的字段名，通过 db tag 为未来的 ORM/映射工具预留元数据。
package models

import (
	"database/sql"
	"time"
)

// User 表示一个注册用户，对应数据库中的 users 表。
//
// 字段说明：
//   - ID: 用户的唯一标识，数据库中使用 UUID 类型并自动生成（gen_random_uuid）。
//   - Email: 用户的邮箱地址，具有唯一约束，用于账号密码登录。（可为 NULL，支持手机号登录）
//   - Phone: 用户手机号，用于短信验证码登录。（可为 NULL，支持邮箱登录）
//   - Password: 经过 bcrypt 哈希后的密码。
//     json tag 为 "-"，确保该字段在序列化响应时不会被暴露给客户端。
//   - CreatedAt: 账户创建时间。
//   - UpdatedAt: 账户最后更新时间。
type User struct {
	ID        string         `json:"id" db:"id"`
	Email     sql.NullString `json:"email" db:"email"`
	Phone     sql.NullString `json:"phone" db:"phone"`
	Password  string         `json:"-" db:"password"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}
