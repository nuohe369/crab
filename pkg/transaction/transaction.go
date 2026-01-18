// Package transaction provides transaction management utilities for database operations
// Package transaction 提供数据库操作的事务管理工具
package transaction

import (
	"context"
	"errors"

	"xorm.io/xorm"
)

// contextKey is used to store session in context
// contextKey 用于在上下文中存储会话
type contextKey string

const sessionKey contextKey = "tx_session" // Session key in context | 上下文中的会话键

// WithTransaction executes function within a transaction (recommended)
// Automatically handles Begin, Commit, Rollback and Close
// WithTransaction 在事务中执行函数（推荐）
// 自动处理 Begin、Commit、Rollback 和 Close
//
// Example:
//
//	err := transaction.WithTransaction(db, func(session *xorm.Session) error {
//	    // Business logic
//	    _, err := session.Insert(&user)
//	    if err != nil {
//	        return err // Auto rollback
//	    }
//	    _, err = session.Insert(&profile)
//	    return err // Auto commit or rollback
//	})
func WithTransaction(db *xorm.Engine, fn func(*xorm.Session) error) error {
	if db == nil {
		return errors.New("db engine is nil")
	}

	session := db.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// Execute business logic | 执行业务逻辑
	if err := fn(session); err != nil {
		session.Rollback()
		return err
	}

	// Commit transaction | 提交事务
	return session.Commit()
}

// WithTxContext executes function within a transaction with context support
// Session is stored in context and can be retrieved via GetSession
// WithTxContext 在事务中执行函数，支持上下文
// 会话存储在上下文中，可通过 GetSession 获取
//
// Example:
//
//	err := transaction.WithTxContext(ctx, db, func(ctx context.Context) error {
//	    session := transaction.GetSession(ctx)
//	    _, err := session.Insert(&user)
//	    return err
//	})
func WithTxContext(ctx context.Context, db *xorm.Engine, fn func(context.Context) error) error {
	if db == nil {
		return errors.New("db engine is nil")
	}

	session := db.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// Store session in context | 将会话存储到上下文
	txCtx := context.WithValue(ctx, sessionKey, session)

	// Execute business logic | 执行业务逻辑
	if err := fn(txCtx); err != nil {
		session.Rollback()
		return err
	}

	// Commit transaction | 提交事务
	return session.Commit()
}

// GetSession retrieves session from context
// GetSession 从上下文中获取会话
func GetSession(ctx context.Context) *xorm.Session {
	if session, ok := ctx.Value(sessionKey).(*xorm.Session); ok {
		return session
	}
	return nil
}

// IsInTransaction checks if context is within a transaction
// IsInTransaction 检查上下文是否在事务中
func IsInTransaction(ctx context.Context) bool {
	return GetSession(ctx) != nil
}

// WithNestedTransaction supports nested transactions
// If already in a transaction, reuses the existing session; otherwise creates a new transaction
// WithNestedTransaction 支持嵌套事务
// 如果已在事务中，则复用现有会话；否则创建新事务
//
// Example:
//
//	err := transaction.WithNestedTransaction(db, session, func(s *xorm.Session) error {
//	    // If session is not nil, reuse it; otherwise create new transaction
//	    _, err := s.Insert(&user)
//	    return err
//	})
func WithNestedTransaction(db *xorm.Engine, existingSession *xorm.Session, fn func(*xorm.Session) error) error {
	// If already in transaction, reuse it | 如果已在事务中，则复用
	if existingSession != nil {
		return fn(existingSession)
	}

	// Otherwise create new transaction | 否则创建新事务
	return WithTransaction(db, fn)
}

// MustTransaction executes function within a transaction, panics on failure
// Suitable for initialization scenarios that must succeed
// MustTransaction 在事务中执行函数，失败时 panic
// 适用于必须成功的初始化场景
func MustTransaction(db *xorm.Engine, fn func(*xorm.Session) error) {
	if err := WithTransaction(db, fn); err != nil {
		panic(err)
	}
}

// TransactionFunc is the transaction function type
// TransactionFunc 是事务函数类型
type TransactionFunc func(*xorm.Session) error

// Chain executes multiple transaction functions in sequence
// All functions execute within the same transaction, any failure triggers rollback
// Chain 按顺序执行多个事务函数
// 所有函数在同一事务中执行，任何失败都会触发回滚
//
// Example:
//
//	err := transaction.Chain(db,
//	    func(s *xorm.Session) error { return createUser(s) },
//	    func(s *xorm.Session) error { return createProfile(s) },
//	    func(s *xorm.Session) error { return sendEmail(s) },
//	)
func Chain(db *xorm.Engine, fns ...TransactionFunc) error {
	return WithTransaction(db, func(session *xorm.Session) error {
		for _, fn := range fns {
			if err := fn(session); err != nil {
				return err
			}
		}
		return nil
	})
}
