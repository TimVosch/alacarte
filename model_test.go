//nolint:errcheck
package alacarte_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pollex.nl/alacarte"
)

var (
	comment = alacarte.New[Comment]("book_comments").
		AddSimpleField("id", func(t *Comment) any { return &t.ID }).
		AddSimpleField("name", func(t *Comment) any { return &t.Name }).
		AddSimpleField("book_id", func(t *Comment) any { return &t.BookID })

	//
	book = alacarte.New[Book]("books").
		AddSimpleField("id", func(t *Book) any { return &t.ID }).
		AddSimpleField("name", func(t *Book) any { return &t.Name }).
		AddSimpleField("author_id", func(t *Book) any { return &t.AuthorID }).
		AddRelation("comments",
			alacarte.HasMany(comment,
				func(book Book, comment Comment) bool { return comment.BookID == book.ID },
				func(book *Book, comments []Comment) { book.Comments = comments },
				alacarte.WhereIDs("book_id", func(book Book) uint64 { return book.ID }),
				alacarte.DependsOn("id", "comments.book_id"),
			),
		)

	//
	author = alacarte.New[Author]("authors").
		AddSimpleField("id", func(t *Author) any { return &t.ID }).
		AddSimpleField("name", func(t *Author) any { return &t.Name }).
		AddField(
			"tags",
			alacarte.Col("tags"),
			func(t *Author) (alacarte.Ptrs, alacarte.Action) {
				var tagString string
				return alacarte.Ptrs{&tagString}, func() {
					t.Tags = strings.Split(tagString, ",")
				}
			},
		).
		AddRelation(
			"books",
			alacarte.HasMany(book,
				func(author Author, book Book) bool { return book.AuthorID == author.ID },
				func(author *Author, books []Book) { author.Books = books },
				alacarte.WhereIDs("author_id", func(a Author) uint64 { return a.ID }),
				alacarte.DependsOn("id", "books.author_id"),
			),
		)
)

func init() {
	comment.AddRelation(
		"book",
		alacarte.HasOne(book,
			func(c Comment, b Book) bool { return c.BookID == b.ID },
			func(c *Comment, b Book) { c.Book = &b },
			alacarte.WhereIDs("id", func(c Comment) uint64 { return c.BookID }),
			alacarte.DependsOn(),
		))
}

func TestBasicModelUsage(t *testing.T) {
	// Arrange
	db, sq := setupDB(t)
	sq.Insert("authors").Values(1, "Jeff", "cool,awesome").
		Values(2, "Madonna", "vocal").Exec()

	t.Run("select fields", func(t *testing.T) {
		authors, err := author.Query("id", "tags").Collect(context.Background(), db)
		require.NoError(t, err)

		// Assert
		assert.Len(t, authors, 2)
		for i := range 2 {
			assert.Empty(t, authors[i].Name)
			assert.NotEmpty(t, authors[i].ID)
			assert.NotEmpty(t, authors[i].Tags)
		}
	})

	t.Run("select all by not providing fields", func(t *testing.T) {
		authors, err := author.Query().Collect(context.Background(), db)
		require.NoError(t, err)

		// Assert
		assert.Len(t, authors, 2)
		for i := range 2 {
			assert.NotEmpty(t, authors[i].Name)
			assert.NotEmpty(t, authors[i].ID)
			assert.NotEmpty(t, authors[i].Tags)
		}
	})
}

func TestBasicModelRelation(t *testing.T) {
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

	t.Run("relation all fields", func(t *testing.T) {
		authors, err := author.Query("id", "books").
			Collect(context.Background(), db)
		require.NoError(t, err)

		require.Len(t, authors, 2)
		require.Len(t, authors[0].Books, 2)
		require.Len(t, authors[1].Books, 2)
		require.NotEmpty(t, authors[0].Books[0].Name)
		require.NotEmpty(t, authors[0].Books[1].Name)
		require.NotEmpty(t, authors[0].Books[0].AuthorID)
		require.NotEmpty(t, authors[0].Books[1].AuthorID)
	})

	t.Run("base and relation all fields", func(t *testing.T) {
		authors, err := author.Query("*", "books").
			Collect(context.Background(), db)
		require.NoError(t, err)

		require.Len(t, authors, 2)
		require.Len(t, authors[0].Books, 2)
		require.Len(t, authors[1].Books, 2)
		require.NotEmpty(t, authors[0].Books[0].Name)
		require.NotEmpty(t, authors[0].Books[1].Name)
		require.NotEmpty(t, authors[0].Books[0].AuthorID)
		require.NotEmpty(t, authors[0].Books[1].AuthorID)
	})

	t.Run("nested relations with specific fields", func(t *testing.T) {
		authors, err := author.Query("id", "name", "books.id", "books.author_id", "books.comments.name", "books.comments.book_id").
			Collect(context.Background(), db)
		require.NoError(t, err)

		// Assert
		require.Len(t, authors, 2)
		require.Len(t, authors[0].Books, 2)
		require.Len(t, authors[1].Books, 2)
		require.Len(t, authors[0].Books[0].Comments, 1)
		require.Len(t, authors[0].Books[1].Comments, 1)
		require.Len(t, authors[1].Books[0].Comments, 1)
		require.Len(t, authors[1].Books[1].Comments, 1)

		require.Empty(t, authors[0].Books[0].Name)
		require.Empty(t, authors[0].Books[0].Comments[0].ID)
	})

	t.Run("backref", func(t *testing.T) {
		books, err := book.Query("*", "comments", "comments.book").
			Collect(context.Background(), db)
		require.NoError(t, err)

		require.Len(t, books, 4)
		for _, book := range books {
			assert.Equal(t, book.ID, book.Comments[0].Book.ID)
		}
	})

	t.Run("automatically select fields required for relation", func(t *testing.T) {
		authors, err := author.Query("books.name").
			Collect(context.Background(), db)
		require.NoError(t, err)

		// Assert
		require.Len(t, authors, 2)
		require.Len(t, authors[0].Books, 2)
		require.Len(t, authors[1].Books, 2)

		require.NotEmpty(t, authors[0].ID)
		require.NotEmpty(t, authors[0].Books[0].AuthorID)
		require.NotEmpty(t, authors[0].Books[0].Name)
		require.Empty(t, authors[0].Books[0].ID)
	})

	t.Run("CollectOne should return one item", func(t *testing.T) {
		author, err := author.Query().
			ModifyQuery(func(q alacarte.Q, table string) alacarte.Q { return q.Where("id = ?", 2) }).
			CollectOne(context.Background(), db)
		require.NoError(t, err)
		assert.NotNil(t, author)
		assert.NotEmpty(t, author.ID)
		assert.NotEmpty(t, author.Name)
		assert.Empty(t, author.Books)
	})

	t.Run("CollectOne should error on many returns", func(t *testing.T) {
		author, err := author.Query().
			CollectOne(context.Background(), db)
		assert.ErrorIs(t, err, alacarte.ErrTooManyResults)
		assert.Nil(t, author)
	})

	t.Run("CollectOne should error on no returns", func(t *testing.T) {
		author, err := author.Query().
			ModifyQuery(func(q alacarte.Q, table string) alacarte.Q { return q.Where("false") }).
			CollectOne(context.Background(), db)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, author)
	})
}
