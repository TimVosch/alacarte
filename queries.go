package alacarte

import (
	"log/slog"
)

func Collect[T any](q Q, scans RowScan[T]) ([]T, error) {
	rows, err := q.Query()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Default().Error("Collect: failed to close rows", "error", err.Error())
		}
	}()

	var collection []T
	for rows.Next() {
		var t T
		pointers, actions := scans(&t)
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		actions()
		collection = append(collection, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return collection, nil
}
