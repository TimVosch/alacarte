package alacarte

import (
	"context"
	"database/sql"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
)

type Model[T any] struct {
	// definition
	Table     string
	Fields    map[string]FieldType[T]
	Relations map[string]Resolve[T]

	// running
	fields            []string
	relations         []string
	relationRelations map[string][]string
	relationFields    map[string][]string
	tableAlias        string
	mods              []QueryMod
}

func NewModel[T any](table string) Model[T] {
	model := Model[T]{
		Table:             table,
		tableAlias:        table,
		Fields:            map[string]FieldType[T]{},
		Relations:         make(map[string]Resolve[T]),
		fields:            []string{},
		relations:         []string{},
		relationFields:    map[string][]string{},
		relationRelations: map[string][]string{},
	}

	return model
}

func (model Model[T]) AddField(name string, mod QueryMod, rowScan RowScan[T]) Model[T] {
	model.Fields[name] = Field(mod, rowScan)

	return model
}

func (model Model[T]) AddFieldType(name string, field FieldType[T]) Model[T] {
	model.Fields[name] = field

	return model
}

// AddSimpleField When the field name is the same as the column name and maps directly, use this.
func (model Model[T]) AddSimpleField(name string, ptr func(t *T) any) Model[T] {
	model = model.AddField(name, Col(name), Ptr(ptr))

	return model
}

func (model Model[T]) AddRelation(name string, relation Resolve[T]) Model[T] {
	model.Relations[name] = relation

	return model
}

func (model Model[T]) Mod(mod QueryMod) Model[T] {
	model.mods = append(model.mods, mod)

	return model
}

func (model Model[T]) Select(fieldNames ...string) Model[T] {
	for _, name := range fieldNames {
		parts := strings.SplitN(name, ".", 2)
		if len(parts) == 1 {
			model.fields = append(model.fields, parts[0])
			continue
		}

		model.relationFields[parts[0]] = append(model.relationFields[parts[0]], parts[1])
	}

	return model
}

func (model Model[T]) Resolve(relations ...string) Model[T] {
	for _, name := range relations {
		parts := strings.SplitN(name, ".", 2)
		if len(parts) == 1 {
			model.relations = append(model.relations, parts[0])
			continue
		}

		model.relationRelations[parts[0]] = append(model.relationRelations[parts[0]], parts[1])
	}

	return model
}

func (model Model[T]) Collect(ctx context.Context, db *sql.DB) ([]T, error) {
	q := squirrel.StatementBuilder.RunWith(db).Select().From(model.Table)

	// Apply runtime mods
	q = applyMods(q, model.tableAlias, model.mods)

	// Remove duplicate fields
	slices.Sort(model.fields)
	model.fields = slices.Compact(model.fields)

	// Collapse fields
	var scans []RowScan[T]
	for _, fieldName := range model.fields {
		field, ok := model.Fields[fieldName]
		if !ok {
			continue
		}

		q = field.Mod(q, model.tableAlias)
		scans = append(scans, field.RowScan)
	}

	parents, err := Collect(q, flattenRowScan(scans))
	if err != nil {
		return nil, err
	}

	// Resolve relations from fields
	// TODO: make this feature configurable
	for relation := range model.relationFields {
		model = model.Resolve(relation)
	}

	// Remove duplicates
	slices.Sort(model.relations)
	model.relations = slices.Compact(model.relations)

	for _, relationName := range model.relations {
		resolve, ok := model.Relations[relationName]
		if !ok {
			continue
		}

		err = resolve(
			ctx,
			db,
			parents,
			model.relationFields[relationName],
			model.relationRelations[relationName],
		)
		if err != nil {
			return nil, err
		}
	}

	return parents, nil
}
