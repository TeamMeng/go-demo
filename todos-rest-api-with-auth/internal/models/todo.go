// package models 定义应用的核心数据结构（领域模型）。
//
// 该包中的结构体被仓库层、处理器层及序列化层共用，
// 通过 json tag 控制对外 API 的字段名，通过 db tag 为未来的 ORM/映射工具预留元数据。
package models

import "time"

// Todo 表示一条待办事项记录，对应数据库中的 todos 表。
//
// 字段说明：
//   - Id: 自增主键，由数据库自动生成。
//   - Title: 待办事项的标题内容。
//   - Completed: 完成状态，false 表示未完成，true 表示已完成。
//   - CreatedAt: 记录创建时间，由数据库默认值 CURRENT_TIMESTAMP 生成。
//   - UpdatedAt: 记录最后更新时间，更新操作时会由数据库自动刷新。
//   - UserID: 外键，关联到 users 表的 id，表示该待办事项属于哪个用户。
type Todo struct {
	Id        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Completed bool      `json:"completed" db:"completed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UserID    string    `json:"user_id" db:"user_id"`
}
