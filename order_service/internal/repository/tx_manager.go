package repository

import (
	"context"
	"database/sql"
	"e-commerce/order_service/internal/domain"
)

type SQLTxManagerImpl struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) domain.TxManager {
	return &SQLTxManagerImpl{db: db}
}

type txKey struct{}

func (m *SQLTxManagerImpl) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx := context.WithValue(ctx, txKey{}, tx)

	err = fn(txCtx)
	if err != nil {
		return err
	}

	return tx.Commit()
}
