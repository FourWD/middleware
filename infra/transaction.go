package infra

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (tm *GormTransactionManager) WithinTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(fn)
}
