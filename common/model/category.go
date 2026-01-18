package model

import (
	"time"

	"github.com/nuohe369/crab/pkg/snowflake"
)

// Category represents the category domain model
// Category 分类领域模型
type Category struct {
	ID        int64     `json:"id" xorm:"pk 'id'"`                      // Category ID | 分类ID
	Name      string    `json:"name" xorm:"varchar(50) notnull 'name'"` // Category name | 分类名称
	Sort      int       `json:"sort" xorm:"default(0) 'sort'"`          // Sort order, smaller values come first | 排序，越小越靠前
	Status    int       `json:"status" xorm:"default(1) 'status'"`      // Status: 1=enabled, 0=disabled | 状态: 1=正常, 0=禁用
	CreatedAt time.Time `json:"created_at" xorm:"created 'created_at'"` // Creation time | 创建时间
	UpdatedAt time.Time `json:"updated_at" xorm:"updated 'updated_at'"` // Update time | 更新时间
}

// TableName returns the table name
// TableName 返回表名
func (c *Category) TableName() string {
	return "category"
}

// DBName specifies the database name
// DBName 指定数据库名称
func (c *Category) DBName() string {
	return "crab_business"
}

// BeforeInsert generates snowflake ID before insertion
// BeforeInsert 插入前生成雪花 ID
func (c *Category) BeforeInsert() {
	if c.ID == 0 {
		c.ID = snowflake.Generate()
	}
}

// IsEnabled checks if the category is enabled
// IsEnabled 检查是否启用
func (c *Category) IsEnabled() bool {
	return c.Status == 1
}

// Enable enables the category
// Enable 启用分类
func (c *Category) Enable() {
	c.Status = 1
}

// Disable disables the category
// Disable 禁用分类
func (c *Category) Disable() {
	c.Status = 0
}

// IsValid validates the category data
// IsValid 校验数据有效性
func (c *Category) IsValid() bool {
	return c.Name != ""
}
