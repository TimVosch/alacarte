package example

import (
	"context"
	"database/sql"
)

type Store struct {
	db *sql.DB
}

type Opts struct {
	Fields  []string
	Expands []string
}

func (store *Store) ListAuthors(ctx context.Context, opts Opts) ([]Author, error) {
	// Expanding a relation is done by `.Query("books")` but `.Query("books.name")` automatically expands it as well.
	// by just appending both fields and expands we expand what we want and select what we want.
	authors, err := AuthorSchema.Query(append(opts.Fields, opts.Expands...)...).Collect(ctx, store.db)
	if err != nil {
		return nil, err
	}

	return authors, nil
}
