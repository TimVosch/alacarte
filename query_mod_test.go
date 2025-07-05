package alacarte_test

import (
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"pollex.nl/alacarte"
)

func TestQueryModColShouldWork(t *testing.T) {
	mod := alacarte.Col("id")
	q := mod(squirrel.Select(), "table")
	queryString, _ := q.MustSql()
	assert.Equal(t, "SELECT table.id", queryString)
}

func TestQueryModColShouldWorkWithMany(t *testing.T) {
	mod := alacarte.Col("id", "name")
	q := mod(squirrel.Select(), "table")
	queryString, _ := q.MustSql()
	assert.Equal(t, "SELECT table.id, table.name", queryString)
}
