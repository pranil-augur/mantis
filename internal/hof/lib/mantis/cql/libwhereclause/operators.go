package libwhereclause

// Operator constants
const (
	// Logical operators
	OpAnd = "and"
	OpOr  = "or"
	OpNot = "not"

	// Comparison operators
	OpEqual    = "="
	OpNotEqual = "!="
	OpRegex    = "=~"
	OpIn       = "in"
	OpContains = "contains"

	// List operators
	OpAny = "any"
	OpAll = "all"
)
