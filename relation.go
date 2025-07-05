package alacarte

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/samber/lo"
)

type (
	Resolve[M any]   func(ctx context.Context, db *sql.DB, parents []M, fields []string) error
	FieldCheck       func(fields string) error
	Binder[M, N any] func(parents []M, children []N)
)

type Relation[M any] struct {
	Resolve Resolve[M]
	Check   FieldCheck
}

func HasMany[M, N any](
	child *ModelSchema[N],
	belongTogether func(M, N) bool,
	assign func(*M, []N),
	wherer func(parents []M) QueryMod,
) Relation[M] {
	return CreateRelation(child, BindBy(belongTogether, assign), wherer)
}

func HasOne[M, N any](
	child *ModelSchema[N],
	belongTogether func(M, N) bool,
	assign func(*M, N),
	wherer func(parents []M) QueryMod,
) Relation[M] {
	return CreateRelation(child, BindByOne(belongTogether, assign), wherer)
}

func CreateRelation[M, N any](
	child *ModelSchema[N],
	binder Binder[M, N],
	wherer func(parents []M) QueryMod,
) Relation[M] {
	return Relation[M]{
		Check: func(field string) error {
			return child.Check(field)
		},
		Resolve: func(ctx context.Context, db *sql.DB, parents []M, fields []string) error {
			children, err := child.Query(fields...).
				ModifyQuery(wherer(parents)).
				Collect(ctx, db)
			if err != nil {
				return err
			}

			binder(parents, children)

			return nil
		},
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

func BindByOne[M, N any](
	belongTogether func(M, N) bool,
	assign func(*M, N),
) Binder[M, N] {
	return func(parents []M, children []N) {
		for ix := range parents {
			parent := &parents[ix]

			for _, child := range children {
				if !belongTogether(*parent, child) {
					continue
				}

				assign(parent, child)
				break
			}

		}
	}
}

func WhereIDs[M any, K any](col string, getID func(m M) K) func(parents []M) QueryMod {
	return func(parents []M) QueryMod {
		return func(q Q, table string) Q {
			return q.Where(
				squirrel.Eq{
					TableCol(table, col): lo.Map(
						parents,
						func(parent M, _ int) K { return getID(parent) },
					),
				},
			)
		}
	}
}
