package alacarte

import (
	"context"
	"database/sql"
)

type (
	Resolve[M any]   func(ctx context.Context, db *sql.DB, parents []M, fields []string, expands []string) error
	Binder[M, N any] func(parents []M, children []N)
)

func HasMany[M, N any](
	child Model[N],
	binder Binder[M, N],
	where func(parents []M) QueryMod,
) Resolve[M] {
	return func(ctx context.Context, db *sql.DB, parents []M, fields, expands []string) error {
		children, err := child.Select(fields...).
			Resolve(expands...).
			Mod(where(parents)).
			Collect(ctx, db)
		if err != nil {
			return err
		}

		binder(parents, children)

		return nil
	}
}

func BindBy[M, N any](
	belongTogether func(M, N) bool,
	assign func(*M, []N),
) Binder[M, N] {
	return func(parents []M, children []N) {
		for ix := range parents {
			parent := &parents[ix]
			var collection []N

			for _, child := range children {
				if !belongTogether(*parent, child) {
					continue
				}

				collection = append(collection, child)
			}

			assign(parent, collection)
		}
	}
}
