package example

import "database/sql"

type Store struct {
	db *sql.DB
}

type Opts struct {
	Fields  []string
	Expands []string
}

func (store *Store) ListAuthors(opts Opts) ([]Author, error) {
	return nil, nil
}
