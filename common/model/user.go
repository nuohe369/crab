package model

import (
	"time"

	"github.com/nuohe369/crab/pkg/snowflake"
	"github.com/nuohe369/crab/pkg/util"
)

// User represents the user domain model
// User 用户领域模型
type User struct {
	ID        int64     `json:"id" xorm:"pk 'id'"`                                     // User ID | 用户ID
	Username  string    `json:"username" xorm:"varchar(50) notnull unique 'username'"` // Username (unique) | 用户名（唯一）
	Nickname  string    `json:"nickname" xorm:"varchar(50) notnull 'nickname'"`        // Nickname | 昵称
	Password  string    `json:"-" xorm:"varchar(255) notnull 'password'"`              // Password (encrypted, not returned in JSON) | 密码（加密，不返回到JSON）
	Status    int       `json:"status" xorm:"default(1) 'status'"`                     // Status: 1=enabled, 0=disabled | 状态: 1=正常, 0=禁用
	CreatedAt time.Time `json:"created_at" xorm:"created 'created_at'"`                // Creation time | 创建时间
	UpdatedAt time.Time `json:"updated_at" xorm:"updated 'updated_at'"`                // Update time | 更新时间
}

// TableName returns the table name
// TableName 返回表名
func (u *User) TableName() string {
	return "user"
}

// DBName specifies the database name
// DBName 指定数据库名称
func (u *User) DBName() string {
	return "crab_usercenter"
}

// BeforeInsert generates snowflake ID before insertion
// BeforeInsert 插入前生成雪花 ID
func (u *User) BeforeInsert() {
	if u.ID == 0 {
		u.ID = snowflake.Generate()
	}
}

// SetPassword sets the password (encrypted)
// SetPassword 设置密码（加密）
func (u *User) SetPassword(password string) error {
	hashed, err := util.HashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hashed
	return nil
}

// CheckPassword validates the password
// CheckPassword 校验密码
func (u *User) CheckPassword(password string) bool {
	return util.CheckPassword(password, u.Password)
}

// IsEnabled checks if the user is enabled
// IsEnabled 检查是否启用
func (u *User) IsEnabled() bool {
	return u.Status == 1
}

// Enable enables the user
// Enable 启用用户
func (u *User) Enable() {
	u.Status = 1
}

// Disable disables the user
// Disable 禁用用户
func (u *User) Disable() {
	u.Status = 0
}

// IsValid validates the user data
// IsValid 校验用户数据有效性
func (u *User) IsValid() bool {
	return u.Username != "" && u.Nickname != ""
}
