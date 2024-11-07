package libwhereclause

import (
	"cuelang.org/go/cue"
)

// Operator field names in CUE
const (
	fieldAnd = "and"
	fieldOr  = "or"
	fieldNot = "not"
	fieldAny = "any"
	fieldAll = "all"
)

// CreateEvaluator creates a WhereEvaluator from a CUE expression
func CreateEvaluator(expr cue.Value) WhereEvaluator {
	if isLogicalOp(expr) {
		return createLogicalEvaluator(expr)
	}

	if isListOp(expr) {
		return createListEvaluator(expr)
	}

	return createComparisonEvaluator(expr)
}

func isLogicalOp(expr cue.Value) bool {
	iter, _ := expr.Fields()
	for iter.Next() {
		switch iter.Label() {
		case fieldAnd, fieldOr, fieldNot:
			return true
		}
	}
	return false
}

func createLogicalEvaluator(expr cue.Value) *LogicalEvaluator {
	iter, _ := expr.Fields()
	for iter.Next() {
		switch iter.Label() {
		case fieldAnd:
			return &LogicalEvaluator{
				Op:    OpAnd,
				Exprs: createEvaluatorList(iter.Value()),
			}
		case fieldOr:
			return &LogicalEvaluator{
				Op:    OpOr,
				Exprs: createEvaluatorList(iter.Value()),
			}
		case fieldNot:
			return &LogicalEvaluator{
				Op:    OpNot,
				Exprs: []WhereEvaluator{CreateEvaluator(iter.Value())},
			}
		}
	}
	return nil
}

func createEvaluatorList(expr cue.Value) []WhereEvaluator {
	var evaluators []WhereEvaluator
	iter, _ := expr.List()
	for iter.Next() {
		evaluators = append(evaluators, CreateEvaluator(iter.Value()))
	}
	return evaluators
}

func isListOp(expr cue.Value) bool {
	iter, _ := expr.Fields()
	for iter.Next() {
		switch iter.Label() {
		case fieldAny, fieldAll:
			return true
		}
	}
	return false
}

func createListEvaluator(expr cue.Value) *ListEvaluator {
	iter, _ := expr.Fields()
	for iter.Next() {
		path := ""
		predicate := CreateEvaluator(iter.Value())

		pathIter, _ := iter.Value().Fields()
		if pathIter.Next() {
			path = pathIter.Label()
		}

		switch iter.Label() {
		case fieldAny:
			return &ListEvaluator{
				Path:      path,
				Operator:  OpAny,
				Predicate: predicate,
			}
		case fieldAll:
			return &ListEvaluator{
				Path:      path,
				Operator:  OpAll,
				Predicate: predicate,
			}
		}
	}
	return nil
}

func createComparisonEvaluator(expr cue.Value) *ComparisonEvaluator {
	iter, _ := expr.Fields()
	if !iter.Next() {
		return nil
	}

	path := iter.Label()
	value := iter.Value()

	// If the value is a list, use OpIn operator
	if _, err := value.List(); err == nil {
		return &ComparisonEvaluator{
			Path:     path,
			Operator: OpIn,
			Expected: value,
		}
	}

	// Default to equality comparison
	return &ComparisonEvaluator{
		Path:     path,
		Operator: OpEqual,
		Expected: value,
	}
}
