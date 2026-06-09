package psql

import (
	"context"

	"gorm.io/gorm"
)

// TransactionManager provides database transaction capabilities.
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes the given function within a database transaction.
// If fn returns an error, the transaction is rolled back. Otherwise, it is committed.
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(fn)
}
