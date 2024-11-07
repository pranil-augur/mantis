package libselectclause

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	types "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/shared"
)

// ProcessClause processes the SELECT clause and populates the QueryResult
func ProcessClause(value cue.Value, selects []string, file string, result *types.QueryResult) {
	for _, sel := range selects {
		if sel == "*" {
			// Handle SELECT *
			if match := ExtractMatch(value, value.Path().String(), file); match != nil {
				result.Matches[sel] = append(result.Matches[sel], *match)
			}
		} else {
			// Handle specific field selection
			fieldValue := value.LookupPath(cue.ParsePath(sel))
			if fieldValue.Exists() {
				if match := ExtractMatch(fieldValue, sel, file); match != nil {
					result.Matches[sel] = append(result.Matches[sel], *match)
				}
			}
		}
	}
}

// ExtractMatch creates a Match object from a CUE value
func ExtractMatch(value cue.Value, path string, file string) *types.Match {
	valueType := getValueType(value)

	match := &types.Match{
		Path:     path,
		File:     file,
		Type:     valueType,
		CueValue: value,
	}

	// Extract the value based on its type
	switch value.Kind() {
	case cue.StringKind:
		if str, err := value.String(); err == nil {
			match.Value = str
		}
	case cue.IntKind:
		if i, err := value.Int64(); err == nil {
			match.Value = fmt.Sprintf("%d", i)
		}
	case cue.FloatKind:
		if f, err := value.Float64(); err == nil {
			match.Value = fmt.Sprintf("%f", f)
		}
	case cue.StructKind:
		var sb strings.Builder
		iter, _ := value.Fields()
		for iter.Next() {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %s", iter.Label(), iter.Value()))

			// Process children recursively
			if childMatch := ExtractMatch(iter.Value(), iter.Label(), file); childMatch != nil {
				match.Children = append(match.Children, childMatch)
			}
		}
		match.Value = sb.String()
	case cue.ListKind:
		var sb strings.Builder
		iter, _ := value.List()
		for iter.Next() {
			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprint(iter.Value()))
		}
		match.Value = "[" + sb.String() + "]"
	}

	return match
}

// FindFieldValue finds a field value in a Match object
func FindFieldValue(match *types.Match, field string) string {
	// Handle wildcard selector
	if field == "*" {
		return match.Value
	}

	if match.Path == field {
		return match.Value
	}

	// Search in children
	for _, child := range match.Children {
		if child.Path == field {
			return child.Value
		}
		// Recursively search in nested structures
		if strings.HasPrefix(field, child.Path+".") {
			if value := FindFieldValue(child, strings.TrimPrefix(field, child.Path+".")); value != "" {
				return value
			}
		}
	}
	return ""
}

// Helper function to get the CUE value type
func getValueType(value cue.Value) string {
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
	case cue.BoolKind:
		return "bool"
	default:
		return "unknown"
	}
}
