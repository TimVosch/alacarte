package alacarte_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"pollex.nl/alacarte"
)

//nolint:errcheck
func TestBasicRelationalModel(t *testing.T) {
	// Arrange
	db, sq := setupDB(t)
	sq.Insert("authors").
		Values(1, "Jeff", "cool,awesome").
		Values(2, "Madonna", "vocal").Exec()
	sq.Insert("books").
		Values(1, "Life of Jeff", 1).
		Values(2, "Cooking like Jeff", 1).
		Values(3, "Sing baby sing", 2).
		Values(4, "the singeth hath endeth", 2).Exec()
	sq.Insert("book_comments").
		Values(1, "Great book!", 1).
		Values(2, "Very insightful", 2).
		Values(3, "A masterpiece", 3).
		Values(4, "Could be better", 4).Exec()

	comment := alacarte.NewModel[Comment]("book_comments").
		AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Comment) any { return &t.ID })).
		AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Comment) any { return &t.Name })).
		AddField("book_id", alacarte.Col("book_id"), alacarte.Ptr(func(t *Comment) any { return &t.BookID }))
	book := alacarte.NewModel[Book]("books").
		AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Book) any { return &t.ID })).
		AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Book) any { return &t.Name })).
		AddField("author_id", alacarte.Col("author_id"), alacarte.Ptr(func(t *Book) any { return &t.AuthorID })).
		AddRelation("comments",
			alacarte.HasMany(comment,
				alacarte.BindBy(
					func(book Book, comment Comment) bool { return comment.BookID == book.ID },
					func(book *Book, comments []Comment) { book.Comments = comments },
				),
				func(books []Book) alacarte.QueryMod {
					return func(q alacarte.Q, table string) alacarte.Q {
						ids := lo.Map(
							books,
							func(book Book, _ int) uint64 { return book.ID },
						)
						return q.Where(squirrel.Eq{"book_comments.book_id": ids})
					}
				},
			),
		)
	author := alacarte.NewModel[Author]("authors").
		AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Author) any { return &t.ID })).
		AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Author) any { return &t.Name })).

		// TODO:  relations should specify mandatory fields required to resolve the relation.
		AddRelation("books",
			alacarte.HasMany(book,
				alacarte.BindBy(
					func(author Author, book Book) bool { return book.AuthorID == author.ID },
					func(author *Author, books []Book) { author.Books = books },
				),
				func(authors []Author) alacarte.QueryMod {
					return func(q alacarte.Q, table string) alacarte.Q {
						ids := lo.Map(
							authors,
							func(author Author, _ int) uint64 { return author.ID },
						)
						return q.Where(squirrel.Eq{"books.author_id": ids})
					}
				},
			),
		)

	authors, err := author.Select("id", "name", "books.id", "books.name", "books.author_id", "books.comments.id", "books.comments.name", "books.comments.book_id").
		Resolve("books", "books.comments").
		Collect(context.Background(), db)
	require.NoError(t, err)

	// Pretty print
	data, _ := json.MarshalIndent(authors, "", "  ")
	t.Logf("%s", string(data))
}
