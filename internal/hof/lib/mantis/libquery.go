/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type QueryConfig struct {
	From   string         `json:"from"`   // Data source path
	Select []string       `json:"select"` // Fields to project
	Where  map[string]any `json:"where"`  // Predicate conditions
}

type QueryResult struct {
	Matches map[string][]Match // key is the expression that matched
}

type Match struct {
	Value    string    // The matched value (string representation)
	CueValue cue.Value // The original CUE value
	Path     string    // CUE path where match was found
	File     string    // File where match was found
	Type     string    // Type of the matched value
	Children []Match   // For nested matches
}

func LoadQueryConfig(path string) (QueryConfig, error) {
	var config QueryConfig

	ctx := cuecontext.New()
	instances := load.Instances([]string{path}, nil)
	if len(instances) == 0 || instances[0].Err != nil {
		return config, fmt.Errorf("failed to load CUE file: %v", instances[0].Err)
	}

	value := ctx.BuildInstance(instances[0])
	if value.Err() != nil {
		return config, fmt.Errorf("failed to build CUE instance: %v", value.Err())
	}

	// Extract FROM path
	if from := value.LookupPath(cue.ParsePath("from")); from.Exists() {
		if str, err := from.String(); err == nil {
			config.From = str
		}
	}

	// Extract SELECT projections
	if selects, err := extractStringSlice(value, "select"); err == nil {
		config.Select = selects
	}

	// Extract WHERE predicates
	config.Where = make(map[string]any)
	if where := value.LookupPath(cue.ParsePath("where")); where.Exists() {
		iter, _ := where.Fields()
		for iter.Next() {
			if val, err := iter.Value().String(); err == nil {
				config.Where[iter.Label()] = val
			}
		}
	}

	return config, nil
}

func QueryConfigurations(directory string, config QueryConfig) (QueryResult, error) {
	result := QueryResult{
		Matches: make(map[string][]Match),
	}

	files, err := getCueFiles(directory)
	if err != nil {
		return result, err
	}

	ctx := cuecontext.New()

	for _, file := range files {
		instances := load.Instances([]string{file}, nil)
		if len(instances) == 0 || instances[0].Err != nil {
			continue
		}

		value := ctx.BuildInstance(instances[0])
		if value.Err() != nil {
			continue
		}

		// First evaluate the FROM clause to get the base value
		baseValue := value
		if config.From != "" {
			if prefix, pattern, suffix, ok := parsePatternExpression(config.From); ok {
				// Handle pattern-based FROM
				matches, err := evaluateFromPattern(value, prefix, pattern, suffix)
				if err != nil {
					continue
				}
				// Process each match with SELECT
				for _, match := range matches {
					processSelectClause(match.CueValue, config.Select, file, config.Where, &result)
				}
			} else {
				// Handle direct path FROM
				baseValue = value.LookupPath(cue.ParsePath(config.From))
				if !baseValue.Exists() {
					continue
				}
				processSelectClause(baseValue, config.Select, file, config.Where, &result)
			}
		}
	}

	return result, nil
}

// New helper function to process SELECT clause
func processSelectClause(value cue.Value, selects []string, file string, filters map[string]any, result *QueryResult) {
	for _, sel := range selects {
		if sel == "*" {
			// Handle SELECT *
			if match := extractMatch(value, value.Path().String(), file, filters); match != nil {
				result.Matches[sel] = append(result.Matches[sel], *match)
			}
		} else {
			// Handle specific field selection
			fieldValue := value.LookupPath(cue.ParsePath(sel))
			if fieldValue.Exists() {
				if match := extractMatch(fieldValue, sel, file, filters); match != nil {
					result.Matches[sel] = append(result.Matches[sel], *match)
				}
			} else {
				// Try to find the field as a direct child of the value
				iter, _ := value.Fields()
				for iter.Next() {
					if childValue := iter.Value().LookupPath(cue.ParsePath(sel)); childValue.Exists() {
						if match := extractMatch(childValue, sel, file, filters); match != nil {
							result.Matches[sel] = append(result.Matches[sel], *match)
						}
					}
				}
			}
		}
	}
}

// New helper function to evaluate FROM patterns
func evaluateFromPattern(value cue.Value, prefix string, pattern string, suffix string) ([]Match, error) {
	rootPath := cue.ParsePath(prefix)
	if rootPath.Err() != nil {
		return nil, rootPath.Err()
	}

	rootValue := value.LookupPath(rootPath)
	if !rootValue.Exists() {
		return nil, nil
	}

	matches := []Match{}
	iter, err := rootValue.Fields()
	if err != nil {
		return nil, err
	}

	for iter.Next() {
		fieldValue := iter.Value()

		if !isMatchingPatternType(fieldValue, pattern) {
			continue
		}

		match := &Match{
			CueValue: fieldValue,
			Path:     iter.Label(),
			Type:     getValueType(fieldValue),
		}

		// Handle different value types
		switch fieldValue.Kind() {
		case cue.StringKind:
			if str, err := fieldValue.String(); err == nil {
				match.Value = str
			}
		case cue.StructKind:
			var sb strings.Builder
			structIter, _ := fieldValue.Fields()
			for structIter.Next() {
				if sb.Len() > 0 {
					sb.WriteString(", ")
				}
				if str, err := structIter.Value().String(); err == nil {
					sb.WriteString(fmt.Sprintf("%s: %s", structIter.Label(), str))
				}
			}
			match.Value = sb.String()
		}
		matches = append(matches, *match)
	}
	return matches, nil
}

func extractMatch(value cue.Value, path string, file string, filters map[string]any) *Match {
	// Check if value matches any of the requested types
	valueType := getValueType(value)
	if !isMatchingValue(value, filters) {
		return nil
	}

	match := &Match{
		Path: path,
		File: file,
		Type: valueType,
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
			fieldValue := iter.Value()
			fieldLabel := iter.Label()

			// Extract the field value based on its kind
			var fieldStr string
			switch fieldValue.Kind() {
			case cue.StringKind:
				if str, err := fieldValue.String(); err == nil {
					fieldStr = str
				}
			case cue.IntKind:
				if i, err := fieldValue.Int64(); err == nil {
					fieldStr = fmt.Sprintf("%d", i)
				}
			default:
				fieldStr = "<complex>"
			}

			if sb.Len() > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %s", fieldLabel, fieldStr))

			// Process children
			if childMatch := extractMatch(fieldValue, fieldLabel, file, filters); childMatch != nil {
				match.Children = append(match.Children, *childMatch)
			}
		}
		match.Value = sb.String()
	}

	return match
}

func FormatQueryResults(result QueryResult, config QueryConfig) string {
	var output strings.Builder

	if len(result.Matches) == 0 {
		return "No matches found in the configurations.\n"
	}

	// Determine fields to display
	var fields []string
	if len(config.Select) == 1 && config.Select[0] == "*" {
		fieldSet := make(map[string]bool)
		for _, matches := range result.Matches {
			for _, match := range matches {
				fieldSet[match.Path] = true
				for _, child := range match.Children {
					fieldSet[match.Path+"."+child.Path] = true
				}
			}
		}
		for field := range fieldSet {
			fields = append(fields, field)
		}
	} else {
		fields = config.Select
	}

	// Print header with file column
	fmt.Fprintf(&output, "%-30s", "file")
	for _, h := range fields {
		fmt.Fprintf(&output, "%-20s", h)
	}
	output.WriteString("\n")

	// Print separator
	output.WriteString(strings.Repeat("-", 30))
	for range fields {
		output.WriteString(strings.Repeat("-", 20))
	}
	output.WriteString("\n")

	// Print values
	for _, matches := range result.Matches {
		for _, match := range matches {
			formatMatchAsTable(&output, match, fields)
		}
	}

	return output.String()
}

func formatMatchAsTable(output *strings.Builder, match Match, fields []string) {
	// Print file name first (showing just the base name for cleaner output)
	fmt.Fprintf(output, "%-30s", filepath.Base(match.File))

	// Print all requested fields
	values := make([]string, len(fields))
	for i, field := range fields {
		value := findFieldValue(match, field)
		values[i] = value
	}

	for _, value := range values {
		fmt.Fprintf(output, "%-20s", value)
	}
	output.WriteString("\n")
}

// Helper functions
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
	default:
		return "unknown"
	}
}

// Add this helper function for pattern matching
func matchPattern(pattern, value string) bool {
	// Handle common glob patterns
	if strings.ContainsAny(pattern, "*?[]") {
		// Convert glob to regex
		regexPattern := globToRegex(pattern)
		matched, err := regexp.MatchString(regexPattern, value)
		return err == nil && matched
	}
	// Fall back to exact match
	return pattern == value
}

// Helper to convert glob patterns to regex
func globToRegex(pattern string) string {
	var regex strings.Builder
	regex.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			regex.WriteString(".*")
		case '?':
			regex.WriteString(".")
		case '[', ']', '(', ')', '+', '.', '^', '$', '|':
			regex.WriteString("\\" + string(pattern[i]))
		default:
			regex.WriteString(string(pattern[i]))
		}
	}

	regex.WriteString("$")
	return regex.String()
}

// Update isMatchingValue to use the new pattern matcher
func isMatchingValue(value cue.Value, filters map[string]any) bool {
	switch value.Kind() {
	case cue.StructKind:
		for filterPath, filterValue := range filters {
			filterStr, ok := filterValue.(string)
			if !ok {
				continue
			}

			matchedValue := value.LookupPath(cue.ParsePath(filterPath))
			if !matchedValue.Exists() {
				return false
			}

			valueStr, err := matchedValue.String()
			if err != nil {
				return false
			}

			if !matchPattern(filterStr, valueStr) {
				return false
			}
		}
		return true

	case cue.StringKind:
		// Handle string values
		for _, filterValue := range filters {
			filterStr, ok := filterValue.(string)
			if !ok {
				continue
			}

			valueStr, err := value.String()
			if err != nil {
				return false
			}

			return matchPattern(filterStr, valueStr)
		}
		return true

	default:
		return false
	}
}

func getCueFiles(directory string) ([]string, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".cue") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func extractStringSlice(value cue.Value, field string) ([]string, error) {
	var result []string
	fieldValue := value.LookupPath(cue.ParsePath(field))
	if !fieldValue.Exists() {
		return result, nil
	}

	iter, err := fieldValue.List()
	if err != nil {
		return nil, fmt.Errorf("field must be a list: %v", err)
	}

	for iter.Next() {
		str, err := iter.Value().String()
		if err != nil {
			return nil, fmt.Errorf("list values must be strings: %v", err)
		}
		result = append(result, str)
	}
	return result, nil
}

// Add helper to parse pattern expressions
func parsePatternExpression(expr string) (prefix string, pattern string, suffix string, ok bool) {
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

func isMatchingPatternType(value cue.Value, pattern string) bool {
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

func findFieldValue(match Match, field string) string {
	// Handle wildcard selector
	if field == "*" {
		// Return all fields as a formatted string or JSON
		return match.Value
	}

	if match.Path == field {
		return match.Value
	}
	for _, child := range match.Children {
		if child.Path == field {
			return child.Value
		}
	}
	return ""
}
