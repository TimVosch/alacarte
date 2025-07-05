package example

import (
	"pollex.nl/alacarte"
)

var DBAuthor = alacarte.NewModel[Author]("authors").
	AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Author) any { return t.ID })).
	AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Author) any { return t.Name })).
	AddRelation("books",
		alacarte.HasMany(
			DBBook,
			alacarte.BindBy(
				func(author Author, book Book) bool { return book.AuthorID == int64(author.ID) },
				func(author *Author, books []Book) { author.Books = books },
			),
			func(parents []Author) alacarte.QueryMod {
				return func(q alacarte.Q, table string) alacarte.Q { return q }
			},
		),
	)

var DBBook = alacarte.NewModel[Book]("books").
	AddField("id", alacarte.Col("id"), alacarte.Ptr(func(t *Book) any { return t.ID })).
	AddField("name", alacarte.Col("name"), alacarte.Ptr(func(t *Book) any { return t.Name }))
