// Package model provides domain models and database operations
// model 包提供领域模型和数据库操作
package model

import (
	"time"

	"github.com/nuohe369/crab/pkg/snowflake"
)

// Article represents the article domain model
// Article 文章领域模型
type Article struct {
	ID         int64     `json:"id" xorm:"pk 'id'"`
	UserID     int64     `json:"user_id" xorm:"notnull index 'user_id'"`         // Author ID (related to crab_usercenter.user) | 作者ID（关联 crab_usercenter.user）
	CategoryID int64     `json:"category_id" xorm:"notnull index 'category_id'"` // Category ID | 分类ID
	Title      string    `json:"title" xorm:"varchar(200) notnull 'title'"`      // Article title | 文章标题
	Content    string    `json:"content" xorm:"text 'content'"`                  // Article content | 文章内容
	ViewCount  int64     `json:"view_count" xorm:"default(0) 'view_count'"`      // View count | 浏览量
	Status     int       `json:"status" xorm:"default(1) 'status'"`              // Status: 1=published, 0=draft, 2=offline | 状态: 1=已发布, 0=草稿, 2=下架
	CreatedAt  time.Time `json:"created_at" xorm:"created 'created_at'"`         // Creation time | 创建时间
	UpdatedAt  time.Time `json:"updated_at" xorm:"updated 'updated_at'"`         // Update time | 更新时间
}

// TableName returns the table name
// TableName 返回表名
func (a *Article) TableName() string {
	return "article"
}

// DBName specifies the database name
// DBName 指定数据库名称
func (a *Article) DBName() string {
	return "crab_business"
}

// BeforeInsert generates snowflake ID before insertion
// BeforeInsert 插入前生成雪花 ID
func (a *Article) BeforeInsert() {
	if a.ID == 0 {
		a.ID = snowflake.Generate()
	}
}

// Article status constants | 文章状态常量
const (
	ArticleStatusDraft     = 0 // Draft | 草稿
	ArticleStatusPublished = 1 // Published | 已发布
	ArticleStatusOffline   = 2 // Offline | 下架
)

// IsDraft checks if the article is a draft
// IsDraft 检查是否为草稿
func (a *Article) IsDraft() bool {
	return a.Status == ArticleStatusDraft
}

// IsPublished checks if the article is published
// IsPublished 检查是否已发布
func (a *Article) IsPublished() bool {
	return a.Status == ArticleStatusPublished
}

// Publish publishes the article
// Publish 发布文章
func (a *Article) Publish() {
	a.Status = ArticleStatusPublished
}

// Offline takes the article offline
// Offline 下架文章
func (a *Article) Offline() {
	a.Status = ArticleStatusOffline
}

// SaveDraft saves the article as a draft
// SaveDraft 保存为草稿
func (a *Article) SaveDraft() {
	a.Status = ArticleStatusDraft
}

// IncrViewCount increments the view count
// IncrViewCount 增加浏览量
func (a *Article) IncrViewCount() {
	a.ViewCount++
}

// IsValid validates the article data
// IsValid 校验数据有效性
func (a *Article) IsValid() bool {
	return a.UserID > 0 && a.CategoryID > 0 && a.Title != ""
}
