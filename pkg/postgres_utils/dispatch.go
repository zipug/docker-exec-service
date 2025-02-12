package postgres_utils

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func Dispatch[T any](ctx context.Context, db *sqlx.DB, query string, args ...interface{}) ([]T, error) {
	var rows *sqlx.Rows
	if args != nil {
		r, err := db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		rows = r
	} else {
		r, err := db.QueryxContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		rows = r
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, err
		}

		result = append(result, t)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return result, nil
}

func DispatchTx[T any](ctx context.Context, tx *sqlx.Tx, query string, args ...interface{}) ([]T, error) {
	var rows *sqlx.Rows
	if args != nil {
		r, err := tx.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		rows = r
	} else {
		r, err := tx.QueryxContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		rows = r
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {
			return nil, err
		}

		result = append(result, t)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return result, nil
}
