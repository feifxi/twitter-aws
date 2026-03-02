package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store wraps the auto-generated Queries with a real database connection,
// enabling transactional operations via ExecTx.
type Store struct {
	*Queries
	conn *sql.DB
}

// NewStore creates a new Store.
func NewStore(conn *sql.DB) *Store {
	return &Store{
		Queries: New(conn),
		conn:    conn,
	}
}

// ExecTx executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (s *Store) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	q := s.Queries.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	return tx.Commit()
}

// ExecTxAfterCommit executes fn inside a transaction and only runs afterCommit
// if the transaction has successfully committed.
func (s *Store) ExecTxAfterCommit(ctx context.Context, fn func(*Queries) error, afterCommit func()) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	q := s.Queries.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	if afterCommit != nil {
		afterCommit()
	}
	return nil
}
