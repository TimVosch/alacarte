package alacarte

import "fmt"

type ModelSchema[T any] struct {
	Table     string
	Fields    map[string]FieldType[T]
	Relations map[string]Relation[T]
	QueryMods []QueryMod
}

func New[T any](table string) *ModelSchema[T] {
	model := &ModelSchema[T]{
		Table:     table,
		Fields:    map[string]FieldType[T]{},
		Relations: make(map[string]Relation[T]),
	}

	return model
}

func (schema *ModelSchema[T]) AddField(
	name string,
	mod QueryMod,
	rowScan RowScan[T],
) *ModelSchema[T] {
	schema.Fields[name] = Field(mod, rowScan)

	return schema
}

func (schema *ModelSchema[T]) AddFieldType(name string, field FieldType[T]) *ModelSchema[T] {
	schema.Fields[name] = field

	return schema
}

// AddSimpleField When the field name is the same as the column name and maps directly, use this.
func (schema *ModelSchema[T]) AddSimpleField(name string, ptr func(t *T) any) *ModelSchema[T] {
	schema = schema.AddField(name, Col(name), Ptr(ptr))

	return schema
}

func (schema *ModelSchema[T]) AddRelation(name string, relation Relation[T]) *ModelSchema[T] {
	schema.Relations[name] = relation

	return schema
}

func (schema *ModelSchema[T]) ModifyQuery(mod QueryMod) *ModelSchema[T] {
	schema.QueryMods = append(schema.QueryMods, mod)

	return schema
}

func (schema *ModelSchema[T]) Query(fields ...string) ModelQuery[T] {
	return newModelQuery(*schema, fields...)
}

func (schema *ModelSchema[T]) Check(field string) error {
	field, rest := isNested(field)

	if field == "" {
		return nil
	}

	if schema.hasRelation(field) {
		if err := schema.Relations[field].Check(rest); err != nil {
			return err
		}
		return nil
	}

	if schema.hasField(field) {
		if rest != "" {
			return fmt.Errorf("%w: %s", ErrNoSuchField, field)
		}
		return nil
	}

	return fmt.Errorf("%w: %s", ErrNoSuchField, field)
}

func (schema *ModelSchema[T]) hasRelation(name string) bool {
	for k := range schema.Relations {
		if k == name {
			return true
		}
	}
	return false
}

func (schema *ModelSchema[T]) hasField(name string) bool {
	for k := range schema.Fields {
		if k == name {
			return true
		}
	}
	return false
}
