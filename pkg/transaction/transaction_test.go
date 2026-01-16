package transaction

import (
	"context"
	"testing"

	"xorm.io/xorm"
)

// TestWithTransaction_NilDB tests nil db case
func TestWithTransaction_NilDB(t *testing.T) {
	err := WithTransaction(nil, func(session *xorm.Session) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for nil db, got nil")
	}
}

// TestWithTxContext_NilDB tests nil db case
func TestWithTxContext_NilDB(t *testing.T) {
	ctx := context.Background()
	err := WithTxContext(ctx, nil, func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for nil db, got nil")
	}
}

// TestGetSession_NoSession tests no session case
func TestGetSession_NoSession(t *testing.T) {
	ctx := context.Background()
	session := GetSession(ctx)
	if session != nil {
		t.Error("expected nil session, got non-nil")
	}
}

// TestIsInTransaction_False tests not in transaction case
func TestIsInTransaction_False(t *testing.T) {
	ctx := context.Background()
	if IsInTransaction(ctx) {
		t.Error("expected false, got true")
	}
}

// TestWithNestedTransaction_NilSession tests nil session case
func TestWithNestedTransaction_NilSession(t *testing.T) {
	// When existingSession is nil and db is also nil, should return error
	err := WithNestedTransaction(nil, nil, func(session *xorm.Session) error {
		return nil
	})
	if err == nil {
		t.Error("expected error for nil db and nil session, got nil")
	}
}

// TestChain_NilDB tests nil db case
func TestChain_NilDB(t *testing.T) {
	err := Chain(nil)
	if err == nil {
		t.Error("expected error for nil db, got nil")
	}
}
