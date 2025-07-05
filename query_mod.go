package alacarte

import (
	"github.com/Masterminds/squirrel"
	"github.com/samber/lo"
)

type (
	Q        = squirrel.SelectBuilder
	QueryMod func(q Q, table string) Q
)

func Col(names ...string) QueryMod {
	return func(q Q, table string) Q {
		columns := lo.Map(names, func(col string, _ int) string { return TableCol(table, col) })
		return q.Columns(columns...)
	}
}

func TableCol(table, name string) string {
	if table == "" {
		return name
	}
	return table + "." + name
}

func applyMods(q Q, table string, mods []QueryMod) Q {
	for _, mod := range mods {
		q = mod(q, table)
	}

	return q
}
