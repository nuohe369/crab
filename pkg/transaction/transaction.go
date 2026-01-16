package transaction

import (
	"context"
	"errors"

	"xorm.io/xorm"
)

// contextKey is used to store session in context
type contextKey string

const sessionKey contextKey = "tx_session"

// WithTransaction executes function within a transaction (recommended)
// Automatically handles Begin, Commit, Rollback and Close
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

	// Execute business logic
	if err := fn(session); err != nil {
		session.Rollback()
		return err
	}

	// Commit transaction
	return session.Commit()
}

// WithTxContext executes function within a transaction with context support
// Session is stored in context and can be retrieved via GetSession
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

	// Store session in context
	txCtx := context.WithValue(ctx, sessionKey, session)

	// Execute business logic
	if err := fn(txCtx); err != nil {
		session.Rollback()
		return err
	}

	// Commit transaction
	return session.Commit()
}

// GetSession retrieves session from context
func GetSession(ctx context.Context) *xorm.Session {
	if session, ok := ctx.Value(sessionKey).(*xorm.Session); ok {
		return session
	}
	return nil
}

// IsInTransaction checks if context is within a transaction
func IsInTransaction(ctx context.Context) bool {
	return GetSession(ctx) != nil
}

// WithNestedTransaction supports nested transactions
// If already in a transaction, reuses the existing session; otherwise creates a new transaction
//
// Example:
//
//	err := transaction.WithNestedTransaction(db, session, func(s *xorm.Session) error {
//	    // If session is not nil, reuse it; otherwise create new transaction
//	    _, err := s.Insert(&user)
//	    return err
//	})
func WithNestedTransaction(db *xorm.Engine, existingSession *xorm.Session, fn func(*xorm.Session) error) error {
	// If already in transaction, reuse it
	if existingSession != nil {
		return fn(existingSession)
	}

	// Otherwise create new transaction
	return WithTransaction(db, fn)
}

// MustTransaction executes function within a transaction, panics on failure
// Suitable for initialization scenarios that must succeed
func MustTransaction(db *xorm.Engine, fn func(*xorm.Session) error) {
	if err := WithTransaction(db, fn); err != nil {
		panic(err)
	}
}

// TransactionFunc is the transaction function type
type TransactionFunc func(*xorm.Session) error

// Chain executes multiple transaction functions in sequence
// All functions execute within the same transaction, any failure triggers rollback
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
