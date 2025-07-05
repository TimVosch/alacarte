package alacarte_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pollex.nl/alacarte"
)

func TestBasicModelUsage(t *testing.T) {
	// Arrange
	db, sq := setupDB(t)
	sq.Insert("authors").Values(1, "Jeff", "cool,awesome").
		Values(2, "Madonna", "vocal").Exec()

	// Act
	author := alacarte.NewModel[Author]("authors").
		AddField(
			"id",
			alacarte.Col("id"),
			alacarte.Ptr(func(t *Author) any { return &t.ID }),
		).
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
		)

	authors, err := author.Select("id", "tags").Collect(context.Background(), db)
	require.NoError(t, err)

	// Assert
	assert.Len(t, authors, 2)
	for i := range 2 {
		assert.Empty(t, authors[i].Name)
		assert.NotEmpty(t, authors[i].ID)
		assert.NotEmpty(t, authors[i].Tags)
	}
}
