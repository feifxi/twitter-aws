package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(*Queries) error) error
	ExecTxAfterCommit(ctx context.Context, fn func(*Queries) error, afterCommit func()) error
}

// SQLStore provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	conn *pgxpool.Pool
}

// NewStore creates a new Store.
func NewStore(conn *pgxpool.Pool) Store {
	return &SQLStore{
		Queries: New(conn),
		conn:    conn,
	}
}

// ExecTx executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (s *SQLStore) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.Queries.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ExecTxAfterCommit executes fn inside a transaction and only runs afterCommit
// if the transaction has successfully committed.
func (s *SQLStore) ExecTxAfterCommit(ctx context.Context, fn func(*Queries) error, afterCommit func()) error {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.Queries.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if afterCommit != nil {
		afterCommit()
	}
	return nil
}
