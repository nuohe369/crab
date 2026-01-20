// Package model provides domain models and database operations
// model 包提供领域模型和数据库操作
package model

import (
	"time"

	"github.com/nuohe369/crab/pkg/snowflake"
)

// ExampleArticle represents the example article model for demonstration
// ExampleArticle 示例文章模型，用于演示
type ExampleArticle struct {
	ID         snowflake.SnowflakeID `json:"id" xorm:"pk 'id' bigint"`
	UserID     snowflake.SnowflakeID `json:"user_id" xorm:"notnull index 'user_id' bigint"`         // Author ID | 作者ID
	CategoryID snowflake.SnowflakeID `json:"category_id" xorm:"notnull index 'category_id' bigint"` // Category ID | 分类ID
	Title      string                `json:"title" xorm:"varchar(200) notnull 'title'"`             // Article title | 文章标题
	Content    string                `json:"content" xorm:"text 'content'"`                         // Article content | 文章内容
	ViewCount  int64                 `json:"view_count" xorm:"default(0) 'view_count'"`             // View count | 浏览量
	Status     int                   `json:"status" xorm:"default(1) 'status'"`                     // Status: 1=published, 0=draft, 2=offline | 状态: 1=已发布, 0=草稿, 2=下架
	CreatedAt  time.Time             `json:"created_at" xorm:"created 'created_at'"`                // Creation time | 创建时间
	UpdatedAt  time.Time             `json:"updated_at" xorm:"updated 'updated_at'"`                // Update time | 更新时间
}

// TableName returns the table name
// TableName 返回表名
func (a *ExampleArticle) TableName() string {
	return "example_article"
}

// BeforeInsert generates snowflake ID before insertion
// BeforeInsert 插入前生成雪花 ID
func (a *ExampleArticle) BeforeInsert() {
	if a.ID.IsZero() {
		a.ID = snowflake.SnowflakeID(snowflake.Generate())
	}
}

// Article status constants | 文章状态常量
const (
	ExampleArticleStatusDraft     = 0 // Draft | 草稿
	ExampleArticleStatusPublished = 1 // Published | 已发布
	ExampleArticleStatusOffline   = 2 // Offline | 下架
)

// IsDraft checks if the article is a draft
// IsDraft 检查是否为草稿
func (a *ExampleArticle) IsDraft() bool {
	return a.Status == ExampleArticleStatusDraft
}

// IsPublished checks if the article is published
// IsPublished 检查是否已发布
func (a *ExampleArticle) IsPublished() bool {
	return a.Status == ExampleArticleStatusPublished
}

// Publish publishes the article
// Publish 发布文章
func (a *ExampleArticle) Publish() {
	a.Status = ExampleArticleStatusPublished
}

// Offline takes the article offline
// Offline 下架文章
func (a *ExampleArticle) Offline() {
	a.Status = ExampleArticleStatusOffline
}

// SaveDraft saves the article as a draft
// SaveDraft 保存为草稿
func (a *ExampleArticle) SaveDraft() {
	a.Status = ExampleArticleStatusDraft
}

// IncrViewCount increments the view count
// IncrViewCount 增加浏览量
func (a *ExampleArticle) IncrViewCount() {
	a.ViewCount++
}

// IsValid validates the article data
// IsValid 校验数据有效性
func (a *ExampleArticle) IsValid() bool {
	return a.UserID.Valid() && a.CategoryID.Valid() && a.Title != ""
}
