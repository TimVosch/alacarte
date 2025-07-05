package alacarte

import "github.com/Masterminds/squirrel"

type (
	Q        = squirrel.SelectBuilder
	QueryMod func(q Q, table string) Q
)

func Col(name string) QueryMod {
	return func(q Q, table string) Q { return q.Column(TableCol(table, name)) }
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
