package libwhereclause

import (
	"cuelang.org/go/cue"
)

// Evaluate checks if the target value matches the where conditions
func Evaluate(whereValue, targetValue cue.Value) bool {
	evaluator := CreateEvaluator(whereValue)
	return evaluator.Evaluate(targetValue)
}

func (l *LogicalEvaluator) Evaluate(value cue.Value) bool {
	switch l.Op {
	case OpAnd:
		for _, expr := range l.Exprs {
			if !expr.Evaluate(value) {
				return false
			}
		}
		return true
	case OpOr:
		for _, expr := range l.Exprs {
			if expr.Evaluate(value) {
				return true
			}
		}
		return false
	case OpNot:
		return !l.Exprs[0].Evaluate(value)
	}
	return false
}

func (c *ComparisonEvaluator) Evaluate(value cue.Value) bool {
	fieldValue := value.LookupPath(cue.ParsePath(c.Path))
	if !fieldValue.Exists() {
		return false
	}

	switch c.Operator {
	case OpEqual:
		return compareEqual(fieldValue, c.Expected)
	case OpRegex:
		return compareRegex(fieldValue, c.Expected)
	case OpIn:
		return compareIn(fieldValue, c.Expected)
	}
	return false
}

func (l *ListEvaluator) Evaluate(value cue.Value) bool {
	fieldValue := value.LookupPath(cue.ParsePath(l.Path))
	if !fieldValue.Exists() || fieldValue.Kind() != cue.ListKind {
		return false
	}

	iter, _ := fieldValue.List()
	switch l.Operator {
	case OpAny:
		for iter.Next() {
			if l.Predicate.Evaluate(iter.Value()) {
				return true
			}
		}
		return false
	case OpAll:
		for iter.Next() {
			if !l.Predicate.Evaluate(iter.Value()) {
				return false
			}
		}
		return true
	}
	return false
}
