package libwhereclause

import (
	"cuelang.org/go/cue"
)

// WhereEvaluator handles different types of where clause evaluations
type WhereEvaluator interface {
	Evaluate(value cue.Value) bool
}

// LogicalEvaluator handles AND, OR, NOT operations
type LogicalEvaluator struct {
	Op    string // "and", "or", "not"
	Exprs []WhereEvaluator
}

// ComparisonEvaluator handles basic comparisons
type ComparisonEvaluator struct {
	Path     string
	Operator string // "=", "!=", "=~", "in", "contains"
	Expected any
}

// ListEvaluator handles array operations
type ListEvaluator struct {
	Path      string
	Operator  string // "any", "all"
	Predicate WhereEvaluator
}
