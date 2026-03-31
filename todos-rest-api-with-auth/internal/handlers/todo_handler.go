// package handlers 包含所有 HTTP 请求处理器（Handler），负责解析请求、调用仓库层并返回响应。
//
// 该层是应用的接入层（Transport Layer），主要职责包括：
//  1. 从 Gin 上下文中提取认证信息（如 user_id）。
//  2. 解析并校验请求参数（JSON 请求体、URL 路径参数）。
//  3. 调用 repository 层完成业务操作。
//  4. 根据操作结果返回对应的 HTTP 状态码和 JSON 响应。
//  5. 处理并转换底层错误为友好的客户端错误信息。
package handlers

import (
	"net/http"
	"strconv"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTodoInput 是创建待办事项时的请求体结构。
//
// 字段说明：
//   - Title: 待办事项标题，必填，通过 binding:"required" 触发 Gin 的非空校验。
//   - Completed: 初始完成状态，可选，默认为 false。
type CreateTodoInput struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

// UpdateTodoInput 是更新待办事项时的请求体结构，所有字段均为可选（指针类型）。
//
// 使用指针类型的原因：
//   - 可以区分客户端"未传该字段"和"传了零值"两种情况。
//   - 若字段为 nil，表示不更新该字段；若不为 nil，则使用指针指向的值覆盖原值。
type UpdateTodoInput struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// CreateTodoHandler 返回一个 Gin HandlerFunc，用于处理"创建待办事项"请求。
//
// 路由：POST /todos（受保护，需 JWT）
//
// 处理流程：
//  1. 从 Gin 上下文中获取 middleware 注入的 user_id。
//  2. 绑定并校验请求 JSON，确保 title 字段存在且非空。
//  3. 调用 repository.CreateTodo 将数据写入数据库。
//  4. 成功返回 HTTP 201 及创建的 Todo 对象；失败返回 500。
//
// 参数：
//   - pool: PostgreSQL 连接池，通过闭包传递给 Handler 使用。
func CreateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中提取当前登录用户的 user_id
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in this context"})
			return
		}
		userID := userIDInterface.(string)

		// 解析并校验请求体
		var input CreateTodoInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 调用仓库层创建记录
		todo, err := repository.CreateTodo(pool, input.Title, input.Completed, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
			return
		}

		// 返回 201 Created 及新创建的资源
		c.JSON(http.StatusCreated, todo)
	}
}

// GetAllTodosHandler 返回一个 Gin HandlerFunc，用于获取当前登录用户的所有待办事项。
//
// 路由：GET /todos（受保护，需 JWT）
//
// 处理流程：
//  1. 从上下文中提取 user_id。
//  2. 调用 repository.GetAllTodos 查询该用户的全部待办事项，按 created_at 倒序排列。
//  3. 成功返回 HTTP 200 及待办事项列表；失败返回 500。
func GetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in this context"})
			return
		}
		userID := userIDInterface.(string)

		todos, err := repository.GetAllTodos(pool, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, todos)
	}
}

// GetToDoByIDHandler 返回一个 Gin HandlerFunc，用于根据 ID 获取单个待办事项。
//
// 路由：GET /todos/:id（受保护，需 JWT）
//
// 处理流程：
//  1. 提取 user_id 和 URL 路径参数 id。
//  2. 将 id 从字符串转换为整数，若转换失败返回 400。
//  3. 调用 repository.GetToDoByID 查询记录。
//  4. 若记录不存在（pgx.ErrNoRows），返回 404；其他错误返回 500。
func GetToDoByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in this context"})
			return
		}
		userID := userIDInterface.(string)

		idStr := c.Param("id")

		// 将路径参数 id 转换为整数类型
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
			return
		}

		todo, err := repository.GetToDoByID(pool, id, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, todo)
	}
}

// UpdateTodoHandler 返回一个 Gin HandlerFunc，用于根据 ID 更新待办事项。
//
// 路由：PUT /todos/:id（受保护，需 JWT）
//
// 处理流程：
//  1. 提取 user_id 和 URL 路径参数 id，并校验 id 合法性。
//  2. 绑定请求 JSON，校验至少提供一个更新字段（title 或 completed）。
//  3. 查询现有记录，确认该待办事项存在且属于当前用户；不存在返回 404。
//  4. 对未提供的字段保持原值不变，对提供的字段进行覆盖。
//  5. 调用 repository.UpdateTodo 执行更新，成功返回 200 及更新后的记录。
//
// 注意：当前实现采用"先查后更"策略，确保部分更新时未传字段不会被清零。
func UpdateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in this context"})
			return
		}
		userID := userIDInterface.(string)

		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
			return
		}

		var input UpdateTodoInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 校验至少有一个字段需要更新
		if input.Title == nil && input.Completed == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field (title or completed) must be provided"})
		}

		// 查询现有记录，用于部分更新时保留未修改字段的原始值
		existing, err := repository.GetToDoByID(pool, id, userID)

		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 若客户端未传 title，则保持原值
		title := existing.Title
		if input.Title != nil {
			title = *input.Title
		}

		// 若客户端未传 completed，则保持原值
		completed := existing.Completed
		if input.Completed != nil {
			completed = *input.Completed
		}

		todo, err := repository.UpdateTodo(pool, id, title, completed, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return

		}

		c.JSON(http.StatusOK, todo)
	}
}

// DeleteTodoHandler 返回一个 Gin HandlerFunc，用于根据 ID 删除待办事项。
//
// 路由：DELETE /todos/:id（受保护，需 JWT）
//
// 处理流程：
//  1. 提取 user_id 和 URL 路径参数 id，并校验 id 合法性。
//  2. 调用 repository.DeleteTodo 执行删除。
//  3. 若记录不存在，返回 404；删除成功返回 200 及成功消息。
func DeleteTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found in this context"})
			return
		}
		userID := userIDInterface.(string)

		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
			return
		}

		err = repository.DeleteTodo(pool, id, userID)
		if err != nil {
			if err.Error() == "todo with id "+idStr+" not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
	}
}
