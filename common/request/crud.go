// Package request provides request data structures
// request 包提供请求数据结构
package request

// ================ User | 用户 ================

// CreateUserReq represents the create user request
// CreateUserReq 创建用户请求
type CreateUserReq struct {
	Username string `json:"username"` // Username | 用户名
	Nickname string `json:"nickname"` // Nickname | 昵称
	Password string `json:"password"` // Password | 密码
}

// UpdateUserReq represents the update user request
// UpdateUserReq 更新用户请求
type UpdateUserReq struct {
	ID       string `json:"id"`       // User ID | 用户ID
	Nickname string `json:"nickname"` // Nickname | 昵称
	Status   *int   `json:"status"`   // Status | 状态
}

// ================ Category | 分类 ================

// CreateCategoryReq represents the create category request
// CreateCategoryReq 创建分类请求
type CreateCategoryReq struct {
	Name string `json:"name"` // Category name | 分类名称
	Sort int    `json:"sort"` // Sort order | 排序
}

// UpdateCategoryReq represents the update category request
// UpdateCategoryReq 更新分类请求
type UpdateCategoryReq struct {
	ID     string `json:"id"`     // Category ID | 分类ID
	Name   string `json:"name"`   // Category name | 分类名称
	Sort   *int   `json:"sort"`   // Sort order | 排序
	Status *int   `json:"status"` // Status | 状态
}

// ================ Article | 文章 ================

// CreateArticleReq represents the create article request
// CreateArticleReq 创建文章请求
type CreateArticleReq struct {
	UserID     string `json:"user_id"`     // User ID | 用户ID
	CategoryID string `json:"category_id"` // Category ID | 分类ID
	Title      string `json:"title"`       // Article title | 文章标题
	Content    string `json:"content"`     // Article content | 文章内容
	Status     int    `json:"status"`      // Status: 0=draft, 1=published | 状态: 0=草稿, 1=发布
}

// UpdateArticleReq represents the update article request
// UpdateArticleReq 更新文章请求
type UpdateArticleReq struct {
	ID         string `json:"id"`          // Article ID | 文章ID
	CategoryID string `json:"category_id"` // Category ID | 分类ID
	Title      string `json:"title"`       // Article title | 文章标题
	Content    string `json:"content"`     // Article content | 文章内容
	Status     *int   `json:"status"`      // Status | 状态
}

// ListArticleReq represents the list articles request
// ListArticleReq 文章列表请求
type ListArticleReq struct {
	PageReq           // Pagination | 分页
	UserID     string `json:"user_id" query:"user_id"`         // User ID filter | 用户ID筛选
	CategoryID string `json:"category_id" query:"category_id"` // Category ID filter | 分类ID筛选
	Status     *int   `json:"status" query:"status"`           // Status filter | 状态筛选
}
