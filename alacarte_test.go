package alacarte_test

import (
	"database/sql"
	"testing"

	"github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

type Author struct {
	ID    uint64
	Name  string
	Tags  []string
	Books []Book
}

type Book struct {
	ID       uint64
	Name     string
	AuthorID uint64
	Comments []Comment
	Author   *Author
}

type Comment struct {
	ID     uint64
	Name   string
	BookID uint64
	Book   *Book
}

func setupDB(t testing.TB) (*sql.DB, squirrel.StatementBuilderType) {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)

	_, err = db.Exec(migrate)
	require.NoError(t, err)

	sq := squirrel.StatementBuilder.RunWith(db)

	return db, sq
}

const migrate = `
	create table authors (
		id integer not null,
		name text not null,
		tags text not null
	);
	create table books (
		id integer not null,
		name text not null,
		author_id integer
	);
	create table book_comments (
		id integer not null,
		name text not null,
		book_id integer
	);
	`
