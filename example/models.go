package example

import (
	"pollex.nl/alacarte"
)

var AuthorSchema = alacarte.New[Author]("authors").
	AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Author) any { return t.ID })).
	AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Author) any { return t.Name })).
	AddRelation("books",
		alacarte.HasMany(
			DBBook,
			func(author Author, book Book) bool { return book.AuthorID == int64(author.ID) },
			func(author *Author, books []Book) { author.Books = books },
			func(parents []Author) alacarte.QueryMod {
				return func(q alacarte.Q, table string) alacarte.Q { return q }
			},
			alacarte.DependsOn("id", "books.author_id"),
		),
	)

var DBBook = alacarte.New[Book]("books").
	AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Book) any { return t.ID })).
	AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Book) any { return t.Name }))
