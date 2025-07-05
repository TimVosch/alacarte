package alacarte

type (
	Ptrs             []any
	RowScan[T any]   func(*T) (Ptrs, Action)
	Action           func()
	FieldType[T any] struct {
		Mod     QueryMod
		RowScan RowScan[T]
	}
)

func Ptr[T any](ptr func(t *T) any) RowScan[T] {
	return func(t *T) (Ptrs, Action) {
		return Ptrs{ptr(t)}, nil
	}
}

func Field[T any](mod QueryMod, scan RowScan[T]) FieldType[T] {
	return FieldType[T]{mod, scan}
}

func flattenRowScan[T any](rowScans []RowScan[T]) RowScan[T] {
	return func(t *T) (Ptrs, Action) {
		var (
			pointers Ptrs
			actions  []Action
		)
		for _, rowScan := range rowScans {
			ptr, action := rowScan(t)
			pointers = append(pointers, ptr...)
			if action != nil {
				actions = append(actions, action)
			}
		}

		return pointers, flattenActions(actions)
	}
}

func flattenActions(actions []Action) Action {
	return func() {
		for _, action := range actions {
			action()
		}
	}
}
