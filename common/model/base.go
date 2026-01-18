package model

import (
	"log"

	"github.com/nuohe369/crab/pkg/pgsql"
	"xorm.io/xorm"
)

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
// Usage | 用法:
//
//	model.GetDB(&user).Insert(&user)              // default database | 默认数据库
//	model.GetDB(&order).Insert(&order)            // automatically uses order's DBName() | 自动使用 order 的 DBName()
//	model.GetDB(&user, "other").Insert(&user)     // temporarily specify database | 临时指定数据库
func GetDB(model any, name ...string) *xorm.Engine {
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
			// Log error instead of panic to avoid service crash
			// 记录错误而不是 panic，避免服务崩溃
			log.Printf("ERROR: database '%s' not found, falling back to default database", dbName)
			// Panic if default database is also unavailable
			// 如果默认数据库也不可用，则 panic
			if client == nil {
				panic("database '" + dbName + "' not found and no default database available")
			}
		}
	}

	if client == nil {
		panic("database not initialized, please check configuration")
	}

	return client.Engine()
}
