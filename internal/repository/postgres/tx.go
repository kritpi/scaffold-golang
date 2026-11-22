package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// WithinTx carries the pgx.Tx in ctx. Every Repo method called with the inner ctx joins it.
func (r *Repo) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return joinOrBeginTx(ctx, r.pool, fn)
}

func joinOrBeginTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	if _, ok := txFrom(ctx); ok {
		return fn(ctx)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return mapErr(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return mapErr(tx.Commit(ctx))
}

func txFrom(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func conn(ctx context.Context, pool *pgxpool.Pool) querier {
	if tx, ok := txFrom(ctx); ok {
		return tx
	}
	return pool
}
