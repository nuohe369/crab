package model

import (
	"errors"
	"log"

	"github.com/nuohe369/crab/pkg/pgsql"
	"xorm.io/xorm"
)

// ErrDatabaseNotFound is returned when database is not found
var ErrDatabaseNotFound = errors.New("database not found")

// ErrDatabaseNotInitialized is returned when database is not initialized
var ErrDatabaseNotInitialized = errors.New("database not initialized, please check configuration")

// DBNamer interface allows models to specify their default database
// DBNamer 接口允许模型指定其默认数据库
// Usage | 用法:
//
//	type Order struct {
//	    ID int64 `json:"id" xorm:"pk 'id'"`
//	}
//	func (o *Order) DBName() string { return "order_db" }
type DBNamer interface {
	DBName() string
}

// GetDB gets the database engine for a model
// GetDB 获取模型对应的数据库引擎
// Priority: parameter specified > Model's DBName() > default database
// 优先级: 参数指定 > Model 的 DBName() > 默认数据库
//
// ⚠️ WARNING: This function panics if database is not found. For runtime safety, use GetDBSafe instead.
// ⚠️ 警告：如果找不到数据库，此函数会 panic。运行时安全请使用 GetDBSafe。
//
// Usage | 用法:
//
//	model.GetDB(&user).Insert(&user)              // default database | 默认数据库
//	model.GetDB(&order).Insert(&order)            // automatically uses order's DBName() | 自动使用 order 的 DBName()
//	model.GetDB(&user, "other").Insert(&user)     // temporarily specify database | 临时指定数据库
func GetDB(model any, name ...string) *xorm.Engine {
	db, err := GetDBSafe(model, name...)
	if err != nil {
		panic(err)
	}
	return db
}

// GetDBSafe gets the database engine for a model (safe version, returns error instead of panic)
// GetDBSafe 获取模型对应的数据库引擎（安全版本，返回错误而不是 panic）
// Priority: parameter specified > Model's DBName() > default database
// 优先级: 参数指定 > Model 的 DBName() > 默认数据库
//
// Usage | 用法:
//
//	db, err := model.GetDBSafe(&user)
//	if err != nil {
//	    return err
//	}
//	db.Insert(&user)
func GetDBSafe(model any, name ...string) (*xorm.Engine, error) {
	var client *pgsql.Client
	var dbName string

	if len(name) > 0 && name[0] != "" {
		dbName = name[0]
		client = pgsql.Get(dbName)
	} else if namer, ok := model.(DBNamer); ok {
		if dn := namer.DBName(); dn != "" {
			dbName = dn
			client = pgsql.Get(dbName)
		}
	}

	// Fallback to default if client is nil
	// 如果 client 为 nil，回退到默认数据库
	if client == nil {
		client = pgsql.Get()
		if dbName != "" {
			// Log error to help debugging
			// 记录错误以帮助调试
			log.Printf("ERROR: database '%s' not found, falling back to default database", dbName)
			// Return error if default database is also unavailable
			// 如果默认数据库也不可用，则返回错误
			if client == nil {
				return nil, ErrDatabaseNotFound
			}
		}
	}

	if client == nil {
		return nil, ErrDatabaseNotInitialized
	}

	return client.Engine(), nil
}
