package libfromclause

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	types "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/shared"
)

func EvaluatePattern(value cue.Value, prefix, pattern, suffix string) ([]types.Match, error) {
	var matches []types.Match
	baseValue := value.LookupPath(cue.ParsePath(prefix))
	if !baseValue.Exists() {
		return nil, fmt.Errorf("prefix path not found: %s", prefix)
	}

	iter, err := baseValue.Fields()
	if err != nil {
		return nil, err
	}

	for iter.Next() {
		fieldValue := iter.Value()
		if !IsMatchingPatternType(fieldValue, pattern) {
			continue
		}

		if suffix != "" {
			fieldValue = fieldValue.LookupPath(cue.ParsePath(suffix))
			if !fieldValue.Exists() {
				continue
			}
		}

		match := &types.Match{
			Value:    iter.Label(),
			CueValue: fieldValue,
			Path:     prefix + "." + iter.Label(),
			Type:     GetValueType(fieldValue),
		}
		matches = append(matches, *match)
	}

	return matches, nil
}

func ParsePatternExpression(expr string) (prefix string, pattern string, suffix string, ok bool) {
	// Match pattern like "service[string].name"
	parts := strings.Split(expr, "[")
	if len(parts) != 2 {
		return "", "", "", false
	}

	prefix = parts[0] // "service"

	// Split "]" and any suffix
	patternParts := strings.Split(parts[1], "]")
	if len(patternParts) != 2 {
		return "", "", "", false
	}

	pattern = patternParts[0] // "string"
	suffix = patternParts[1]  // ".name"
	if strings.HasPrefix(suffix, ".") {
		suffix = suffix[1:] // remove leading dot
	}

	return prefix, pattern, suffix, true
}

func IsMatchingPatternType(value cue.Value, pattern string) bool {
	switch pattern {
	case "string":
		// Allow struct type when we're matching service entries
		return value.Kind() == cue.StringKind || value.Kind() == cue.StructKind
	case "int":
		return value.Kind() == cue.IntKind
	case "float":
		return value.Kind() == cue.FloatKind
	case "number":
		return value.Kind() == cue.IntKind || value.Kind() == cue.FloatKind
	case "bool":
		return value.Kind() == cue.BoolKind
	case "struct":
		return value.Kind() == cue.StructKind
	case "list":
		return value.Kind() == cue.ListKind
	case "_", "any":
		return true
	default:
		return false
	}
}

func GetValueType(value cue.Value) string {
	switch value.Kind() {
	case cue.StructKind:
		return "struct"
	case cue.StringKind:
		return "string"
	case cue.IntKind:
		return "int"
	case cue.FloatKind:
		return "float"
	case cue.ListKind:
		return "list"
	default:
		return "unknown"
	}
}
