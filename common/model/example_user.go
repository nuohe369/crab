package model

import (
	"time"

	"github.com/nuohe369/crab/pkg/snowflake"
	"github.com/nuohe369/crab/pkg/util"
)

// ExampleUser represents the example user model for demonstration
// ExampleUser 示例用户模型，用于演示
type ExampleUser struct {
	ID        snowflake.SnowflakeID `json:"id" xorm:"pk 'id' bigint"`                      // User ID | 用户ID
	Username  string                `json:"username" xorm:"varchar(50) notnull unique 'username'"` // Username (unique) | 用户名（唯一）
	Nickname  string                `json:"nickname" xorm:"varchar(50) notnull 'nickname'"`        // Nickname | 昵称
	Password  string                `json:"-" xorm:"varchar(255) notnull 'password'"`              // Password (encrypted, not returned in JSON) | 密码（加密，不返回到JSON）
	Status    int                   `json:"status" xorm:"default(1) 'status'"`                     // Status: 1=enabled, 0=disabled | 状态: 1=正常, 0=禁用
	CreatedAt time.Time             `json:"created_at" xorm:"created 'created_at'"`                // Creation time | 创建时间
	UpdatedAt time.Time             `json:"updated_at" xorm:"updated 'updated_at'"`                // Update time | 更新时间
}

// TableName returns the table name
// TableName 返回表名
func (u *ExampleUser) TableName() string {
	return "example_user"
}

// BeforeInsert generates snowflake ID before insertion
// BeforeInsert 插入前生成雪花 ID
func (u *ExampleUser) BeforeInsert() {
	if u.ID.IsZero() {
		u.ID = snowflake.SnowflakeID(snowflake.Generate())
	}
}

// SetPassword sets the password (encrypted)
// SetPassword 设置密码（加密）
func (u *ExampleUser) SetPassword(password string) error {
	hashed, err := util.HashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hashed
	return nil
}

// CheckPassword validates the password
// CheckPassword 校验密码
func (u *ExampleUser) CheckPassword(password string) bool {
	return util.CheckPassword(password, u.Password)
}

// IsEnabled checks if the user is enabled
// IsEnabled 检查是否启用
func (u *ExampleUser) IsEnabled() bool {
	return u.Status == 1
}

// Enable enables the user
// Enable 启用用户
func (u *ExampleUser) Enable() {
	u.Status = 1
}

// Disable disables the user
// Disable 禁用用户
func (u *ExampleUser) Disable() {
	u.Status = 0
}

// IsValid validates the user data
// IsValid 校验用户数据有效性
func (u *ExampleUser) IsValid() bool {
	return u.Username != "" && u.Nickname != ""
}
