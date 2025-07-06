package alacarte

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
)

var (
	// ErrNoSuchField is returned when there is no field or no relation with that name.
	ErrNoSuchField = errors.New("field does not exist")
	// ErrNoSuchRelation is returned only when trying to select a nested field on a relation that does not exist.
	ErrNoSuchRelation = errors.New("relation does not exist")
	// ErrTooManyResults is returned when CollectOne is called but returned many models
	ErrTooManyResults = errors.New("too many result for CollectOne")
)

type ModelQuery[T any] struct {
	schema ModelSchema[T]

	selectedFields         map[string]FieldType[T]
	selectedRelations      map[string]Relation[T]
	selectedRelationFields map[string][]string
	tableAlias             string
	queryMods              []QueryMod

	errors []error
}

func newModelQuery[T any](schema ModelSchema[T], fields ...string) ModelQuery[T] {
	query := ModelQuery[T]{
		schema:                 schema,
		selectedFields:         map[string]FieldType[T]{},
		selectedRelations:      map[string]Relation[T]{},
		selectedRelationFields: map[string][]string{},
		tableAlias:             schema.Table,
		queryMods:              []QueryMod{},
		errors:                 []error{},
	}

	return query.Select(fields...)
}

func (model ModelQuery[T]) ModifyQuery(mod QueryMod) ModelQuery[T] {
	model.queryMods = append(model.queryMods, mod)

	return model
}

func (model ModelQuery[T]) Select(fieldNames ...string) ModelQuery[T] {
	if len(fieldNames) == 0 {
		model.selectAllFields()
		return model
	}

	for _, name := range fieldNames {
		model.resolveSelect(name)
	}

	return model
}

func (model *ModelQuery[T]) resolveSelect(name string) {
	field, rest := isNested(name)

	if field == "*" {
		if rest != "" {
			model.addError(fmt.Errorf("%w: %s", ErrNoSuchRelation, field))
			return
		}

		model.selectAllFields()
		return
	}

	if model.schema.hasRelation(field) {
		if rest != "" && rest != "*" {
			// Validate the chosen nested field.
			if err := model.schema.Relations[field].Check(rest); err != nil {
				model.addError(err)
				return
			}
		}
		model.selectRelation(field, rest)
		return
	}

	if model.schema.hasField(field) {
		// Fields cannot have nesting
		if rest != "" {
			model.addError(fmt.Errorf("%w: %s", ErrNoSuchRelation, field))
			return
		}
		model.selectField(field)
		return
	}

	// Error
	model.addError(fmt.Errorf("%w: %s", ErrNoSuchField, field))
}

func (model *ModelQuery[T]) selectAllFields() {
	for name := range model.schema.Fields {
		model.selectedFields[name] = model.schema.Fields[name]
	}
}

func (model *ModelQuery[T]) selectField(name string) {
	model.selectedFields[name] = model.schema.Fields[name]
}

func (model *ModelQuery[T]) selectRelation(relName, relField string) {
	if relField == "" {
		relField = "*"
	}

	model.selectedRelations[relName] = model.schema.Relations[relName]

	if model.selectedRelationFields[relName] == nil {
		model.selectedRelationFields[relName] = []string{}
	}

	model.selectedRelationFields[relName] = append(model.selectedRelationFields[relName], relField)
}

// =================
// Finishers
// =================

func (model ModelQuery[T]) Err() error {
	return errors.Join(model.errors...)
}

func (model ModelQuery[T]) Collect(ctx context.Context, db squirrel.BaseRunner) ([]T, error) {
	if err := model.Err(); err != nil {
		return nil, err
	}

	parents, err := model.collectBaseModels(ctx, db)
	if err != nil {
		return nil, err
	}

	if err := model.resolveRelations(ctx, db, parents); err != nil {
		return nil, err
	}

	return parents, nil
}

func (model ModelQuery[T]) CollectOne(ctx context.Context, db squirrel.BaseRunner) (*T, error) {
	if err := model.Err(); err != nil {
		return nil, err
	}

	parents, err := model.collectBaseModels(ctx, db)
	if err != nil {
		return nil, err
	}

	if len(parents) == 0 {
		return nil, sql.ErrNoRows
	} else if len(parents) > 1 {
		return nil, ErrTooManyResults
	}

	if err := model.resolveRelations(ctx, db, parents); err != nil {
		return nil, err
	}

	return &parents[0], nil
}

func (model ModelQuery[T]) collectBaseModels(
	ctx context.Context,
	db squirrel.BaseRunner,
) ([]T, error) {
	q := squirrel.StatementBuilder.RunWith(db).Select().From(model.schema.Table)

	// Apply schema mods
	q = applyMods(q, model.tableAlias, model.schema.QueryMods)
	// Apply runtime mods
	q = applyMods(q, model.tableAlias, model.queryMods)

	// Add relation field dependencies
	for _, rel := range model.selectedRelations {
		model = rel.ModelQueryMod(model)
	}

	// Collapse fields
	var scans []RowScan[T]
	for _, field := range model.selectedFields {
		q = field.Mod(q, model.tableAlias)
		scans = append(scans, field.RowScan)
	}

	// Execute query
	parents, err := Collect(ctx, q, flattenRowScan(scans))
	if err != nil {
		return nil, err
	}

	return parents, nil
}

func (model ModelQuery[T]) resolveRelations(
	ctx context.Context,
	db squirrel.BaseRunner,
	parents []T,
) error {
	// Resolve relations
	for name, relation := range model.selectedRelations {
		err := relation.Resolve(
			ctx,
			db,
			parents,
			model.selectedRelationFields[name],
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// =================
// Utilities
// =================

func (model *ModelQuery[T]) addError(err error) {
	model.errors = append(model.errors, err)
}

func isNested(name string) (string, string) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 1 {
		return name, ""
	}
	return parts[0], parts[1]
}
