package postgres

import (
	"github.com/jackc/pgx/v5"
)

func scanRows[T any](
	rows pgx.Rows,
	scanner func(rows pgx.Rows) (T, error)) ([]T, error) {
	defer rows.Close()
	var results []T
	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}
