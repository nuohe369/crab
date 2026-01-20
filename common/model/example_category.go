package model

import (
	"time"

	"github.com/nuohe369/crab/pkg/snowflake"
)

// ExampleCategory represents the example category model for demonstration
// ExampleCategory 示例分类模型，用于演示
type ExampleCategory struct {
	ID        snowflake.SnowflakeID `json:"id" xorm:"pk 'id' bigint"`           // Category ID | 分类ID
	Name      string                `json:"name" xorm:"varchar(50) notnull 'name'"` // Category name | 分类名称
	Sort      int                   `json:"sort" xorm:"default(0) 'sort'"`          // Sort order, smaller values come first | 排序，越小越靠前
	Status    int                   `json:"status" xorm:"default(1) 'status'"`      // Status: 1=enabled, 0=disabled | 状态: 1=正常, 0=禁用
	CreatedAt time.Time             `json:"created_at" xorm:"created 'created_at'"` // Creation time | 创建时间
	UpdatedAt time.Time             `json:"updated_at" xorm:"updated 'updated_at'"` // Update time | 更新时间
}

// TableName returns the table name
// TableName 返回表名
func (c *ExampleCategory) TableName() string {
	return "example_category"
}

// BeforeInsert generates snowflake ID before insertion
// BeforeInsert 插入前生成雪花 ID
func (c *ExampleCategory) BeforeInsert() {
	if c.ID.IsZero() {
		c.ID = snowflake.SnowflakeID(snowflake.Generate())
	}
}

// IsEnabled checks if the category is enabled
// IsEnabled 检查是否启用
func (c *ExampleCategory) IsEnabled() bool {
	return c.Status == 1
}

// Enable enables the category
// Enable 启用分类
func (c *ExampleCategory) Enable() {
	c.Status = 1
}

// Disable disables the category
// Disable 禁用分类
func (c *ExampleCategory) Disable() {
	c.Status = 0
}

// IsValid validates the category data
// IsValid 校验数据有效性
func (c *ExampleCategory) IsValid() bool {
	return c.Name != ""
}
