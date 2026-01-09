package postgres

import (
	"io"

	"github.com/jackc/pgx/v5"
)

type Iterator[T any] interface {
	Next() (T, error)
	Close() error
}

type pgxIterator[T any] struct {
	rows    pgx.Rows
	scanner func(rows pgx.Rows) (T, error)
}

func (it *pgxIterator[T]) Next() (T, error) {
	var zero T
	if !it.rows.Next() {
		return zero, io.EOF
	}
	return it.scanner(it.rows)
}

func (it *pgxIterator[T]) Close() error {
	it.rows.Close()
	return nil
}

func NewIterator[T any](
	rows pgx.Rows,
	scanner func(rows pgx.Rows) (T, error),
) Iterator[T] {
	return &pgxIterator[T]{
		rows:    rows,
		scanner: scanner,
	}
}
