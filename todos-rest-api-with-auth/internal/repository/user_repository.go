// package repository 封装所有与数据库直接交互的操作，为上层（handlers）提供数据持久化能力。
//
// 该层是应用的数据访问层（Data Access Layer），所有函数均接收 *pgxpool.Pool 作为数据库连接，
// 并在内部创建带 5 秒超时的 context.Context，防止因网络或数据库异常导致请求无限期挂起。
package repository

import (
	"context"
	"time"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser 在数据库的 users 表中创建一个新用户，并将数据库生成的字段回写到传入的 user 对象中。
//
// 参数：
//   - pool: PostgreSQL 连接池，用于执行 SQL。
//   - user: 用户对象指针，必须包含 Email 和 Password（已哈希）字段。
//     函数执行成功后，该指针指向的对象会被更新为包含 ID、CreatedAt、UpdatedAt 的完整记录。
//
// 返回值：
//   - *models.User: 创建成功的用户对象（与传入的 user 为同一指针）。
//   - error: 若插入失败（如邮箱唯一约束冲突、连接超时等）则返回错误。
//
// SQL 说明：
//
//	使用 INSERT ... RETURNING 语句一次性完成插入和字段回读，避免二次查询。
func CreateUser(pool *pgxpool.Pool, user *models.User) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id, email, password, created_at, updated_at
	`

	err := pool.QueryRow(ctx, query, user.Email, user.Password).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail 根据邮箱地址查询用户信息。
//
// 参数：
//   - pool: PostgreSQL 连接池。
//   - email: 要查询的邮箱地址。
//
// 返回值：
//   - *models.User: 查询到的用户对象，包含密码哈希字段（供登录时比对）。
//   - error: 若用户不存在返回 pgx.ErrNoRows；其他情况返回对应的数据库错误。
//
// 使用场景：
//
//	主要用于登录流程中根据用户输入的邮箱查找对应账户。
func GetUserByEmail(pool *pgxpool.Pool, email string) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = `
		SELECT id, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID 根据用户 ID 查询用户信息。
//
// 参数：
//   - pool: PostgreSQL 连接池。
//   - id: 用户 ID。
//
// 返回值：
//   - *models.User: 查询到的用户对象。
//   - error: 若用户不存在返回 pgx.ErrNoRows；其他情况返回对应的数据库错误。
//
// 注意：
//
//	当前数据库中 users 表的 id 字段类型为 UUID（字符串），但本函数签名仍使用 int 类型。
//	这会导致传入 UUID 字符串时编译通过但运行时 SQL 类型不匹配。
//	建议后续将参数类型从 int 修改为 string，以与数据库 schema 保持一致。
func GetUserByID(pool *pgxpool.Pool, id int) (*models.User, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query = `
		SELECT id, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
