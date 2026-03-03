package db

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping db integration tests")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func TestExecTxAfterCommitCallsCallbackOnSuccess(t *testing.T) {
	t.Parallel()

	pool := openTestPool(t)
	store := NewStore(pool)

	called := false
	err := store.ExecTxAfterCommit(context.Background(), func(_ *Queries) error {
		return nil
	}, func() {
		called = true
	})
	if err != nil {
		t.Fatalf("ExecTxAfterCommit returned error: %v", err)
	}
	if !called {
		t.Fatal("expected afterCommit callback to be called")
	}
}

func TestExecTxAfterCommitSkipsCallbackOnFailure(t *testing.T) {
	t.Parallel()

	pool := openTestPool(t)
	store := NewStore(pool)

	called := false
	expectedErr := errors.New("tx failure")
	err := store.ExecTxAfterCommit(context.Background(), func(_ *Queries) error {
		return expectedErr
	}, func() {
		called = true
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if called {
		t.Fatal("did not expect afterCommit callback on failed transaction")
	}
}

func TestExecTxReturnsCallbackError(t *testing.T) {
	t.Parallel()

	pool := openTestPool(t)
	store := NewStore(pool)

	expectedErr := errors.New("exec tx failure")
	err := store.ExecTx(context.Background(), func(_ *Queries) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
