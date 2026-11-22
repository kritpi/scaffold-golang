// Package postgres implements port.Repo on PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Config selects the database. An empty DSN selects memory.Repo.
type Config struct {
	DSN      string
	MaxConns int32
}

// Open builds a pool from cfg and pings it.
func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	if cfg.MaxConns > 0 {
		pc.MaxConns = cfg.MaxConns
	}
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

// Repo implements port.Repo on a pgx pool.
type Repo struct {
	pool *pgxpool.Pool
}

// New wraps an open pool.
func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

const (
	classConnectionException   = "08"
	classInsufficientResources = "53"
	classOperatorIntervention  = "57"

	uniqueViolation           = "23505"
	checkViolation            = "23514"
	notNullViolation          = "23502"
	stringDataRightTruncation = "22001"
)

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code[:2] {
		case classConnectionException, classInsufficientResources, classOperatorIntervention:
			return fmt.Errorf("%w: postgres: %v", domain.ErrUnavailable, err)
		}
		switch pgErr.Code {
		case uniqueViolation:
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgErr.ConstraintName)
		case checkViolation, notNullViolation, stringDataRightTruncation:
			return fmt.Errorf("%w: %s", domain.ErrInvalidInput, pgErr.ConstraintName)
		}
		return err
	}
	var connErr *pgconn.ConnectError
	var netErr net.Error
	if errors.As(err, &connErr) || errors.As(err, &netErr) || pgconn.Timeout(err) {
		return fmt.Errorf("%w: postgres: %v", domain.ErrUnavailable, err)
	}
	return err
}

type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
