package example

type Author struct {
	ID   uint64
	Name string

	Books []Book
}

type Book struct {
	ID   uint64
	Name string

	AuthorID int64
	Author   Author
	GenreID  int64
	Genre    Genre
}

type Genre struct {
	ID   uint64
	Name string

	Books []Book
}
