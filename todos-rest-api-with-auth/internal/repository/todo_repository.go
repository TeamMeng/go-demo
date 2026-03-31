// package repository 封装所有与数据库直接交互的操作，为上层（handlers）提供数据持久化能力。
//
// 该层是应用的数据访问层（Data Access Layer），所有函数均接收 *pgxpool.Pool 作为数据库连接，
// 并在内部创建带 5 秒超时的 context.Context，防止因网络或数据库异常导致请求无限期挂起。
package repository

import (
	"context"
	"fmt"
	"time"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTodo 在数据库的 todos 表中创建一个新的待办事项。
//
// 参数：
//   - pool: PostgreSQL 数据库连接池。
//   - title: 待办事项的标题。
//   - completed: 待办事项的初始完成状态。
//   - userID: 所属用户的 ID（外键），用于数据隔离。
//
// 返回值：
//   - *models.Todo: 创建的待办事项对象，包含数据库自动生成的 id、created_at、updated_at。
//   - error: 若插入失败（如外键约束冲突、连接超时等）则返回错误。
//
// SQL 说明：
//
//	使用 INSERT ... RETURNING 一次性完成插入和字段回读，避免二次查询。
func CreateTodo(pool *pgxpool.Pool, title string, completed bool, userID string) (*models.Todo, error) {
	// 创建一个 5 秒超时的上下文，防止数据库操作无限期阻塞
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// SQL 插入语句：插入标题、完成状态和所属用户 ID，并返回自动生成的字段
	var query string = `
		INSERT INTO todos (title, completed, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, completed, created_at, updated_at, user_id
	`

	// 执行查询并将结果扫描到 todo 结构体中
	var todo models.Todo
	var err error = pool.QueryRow(ctx, query, title, completed, userID).Scan(
		&todo.Id,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

// GetAllTodos 从数据库中获取指定用户的所有待办事项，按创建时间倒序排列。
//
// 参数：
//   - pool: PostgreSQL 数据库连接池。
//   - userID: 所属用户的 ID，用于过滤该用户的数据。
//
// 返回值：
//   - []models.Todo: 待办事项列表，按 created_at DESC 排序。
//   - error: 若查询失败则返回错误。
//
// 注意：
//
//	函数内部使用 defer rows.Close() 确保结果集及时关闭，防止连接泄漏。
//	若查询结果为空，返回空切片（长度为 0）而非 nil，方便上层直接序列化为 JSON 数组 []。
func GetAllTodos(pool *pgxpool.Pool, userID string) ([]models.Todo, error) {
	// 创建一个 5 秒超时的上下文
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// SQL 查询语句：获取指定用户的所有待办事项，按创建时间倒序排列
	var query string = `
		SELECT id, title, completed, created_at, updated_at, user_id
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	// 执行查询
	var rows, err = pool.Query(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	// 确保在函数返回前关闭结果集，释放数据库连接资源
	defer rows.Close()

	// 初始化空的待办事项切片（返回空切片而非 nil）
	var todos []models.Todo = []models.Todo{}

	// 遍历查询结果集
	for rows.Next() {
		var todo models.Todo

		// 将当前行的数据扫描到 todo 结构体中
		var err error = rows.Scan(
			&todo.Id,
			&todo.Title,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&todo.UserID,
		)
		if err != nil {
			return nil, err
		}

		// 将扫描后的 todo 添加到结果切片中
		todos = append(todos, todo)
	}

	// 检查遍历过程中是否发生错误（如连接中断等）
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

// GetToDoByID 根据待办事项 ID 和所属用户 ID 查询单条待办事项。
//
// 参数：
//   - pool: PostgreSQL 数据库连接池。
//   - id: 待办事项的自增主键。
//   - userID: 所属用户的 ID，确保用户只能访问自己的数据。
//
// 返回值：
//   - *models.Todo: 查询到的待办事项对象。
//   - error: 若记录不存在返回 pgx.ErrNoRows；其他情况返回对应的数据库错误。
func GetToDoByID(pool *pgxpool.Pool, id int, userID string) (*models.Todo, error) {
	// 创建一个 5 秒超时的上下文
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
		SELECT id, title, completed, created_at, updated_at, user_id
		FROM todos
		WHERE id = $1 AND user_id = $2
	`

	var todo models.Todo

	var err error = pool.QueryRow(ctx, query, id, userID).Scan(
		&todo.Id,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// UpdateTodo 更新指定待办事项的标题和完成状态，并返回更新后的完整记录。
//
// 参数：
//   - pool: PostgreSQL 数据库连接池。
//   - id: 待办事项的自增主键。
//   - title: 新的标题。
//   - completed: 新的完成状态。
//   - userID: 所属用户的 ID，用于权限校验（只能更新自己的记录）。
//
// 返回值：
//   - *models.Todo: 更新后的待办事项对象。
//   - error: 若记录不存在或更新失败则返回错误。
//
// SQL 说明：
//
//	使用 UPDATE ... RETURNING 一次性完成更新和字段回读，同时由数据库自动刷新 updated_at 为 NOW()。
func UpdateTodo(pool *pgxpool.Pool, id int, title string, completed bool, userID string) (*models.Todo, error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
		UPDATE todos
		SET title = $1,
			completed = $2,
			updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, title, completed, created_at, updated_at, user_id
	`

	var todo models.Todo

	err := pool.QueryRow(ctx, query, title, completed, id, userID).Scan(
		&todo.Id,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// DeleteTodo 根据 ID 和所属用户 ID 删除待办事项。
//
// 参数：
//   - pool: PostgreSQL 数据库连接池。
//   - id: 待办事项的自增主键。
//   - userID: 所属用户的 ID，用于权限校验（只能删除自己的记录）。
//
// 返回值：
//   - error: 若删除成功返回 nil；若记录不存在返回 fmt.Errorf("todo with id %d not found", id)；
//     若数据库执行出错则返回对应的数据库错误。
//
// 说明：
//
//	通过 pgx.CommandTag.RowsAffected() 判断实际删除的行数，若为 0 则说明该记录不存在或不属于当前用户。
func DeleteTodo(pool *pgxpool.Pool, id int, userID string) error {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string = `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
	`

	commandTag, err := pool.Exec(ctx, query, id, userID)

	if err != nil {
		return err
	}

	// 若没有任何行被删除，说明该待办事项不存在或当前用户无权删除
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("todo with id %d not found", id)
	}

	return nil
}
