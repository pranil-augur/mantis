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
	Expressions []string       `json:"expressions"` // CUE path expressions
	Filters     map[string]any `json:"filters"`     // Changed to 'any' to support both string and []string
}

type QueryResult struct {
	Matches map[string][]Match // key is the expression that matched
}

type Match struct {
	Value    string  // The matched value
	Path     string  // CUE path where match was found
	File     string  // File where match was found
	Type     string  // Type of the matched value
	Children []Match // For nested matches
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

	// Extract expressions
	if exprs, err := extractStringSlice(value, "expressions"); err == nil {
		config.Expressions = exprs
	}

	// Extract filters
	config.Filters = make(map[string]any)
	if filters := value.LookupPath(cue.ParsePath("filters")); filters.Exists() {
		iter, _ := filters.Fields()
		for iter.Next() {
			if val, err := iter.Value().String(); err == nil {
				config.Filters[iter.Label()] = val
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

		// Process each expression
		for _, expr := range config.Expressions {
			matches, err := evaluateExpression(value, expr, file, config.Filters)
			if err != nil {
				continue
			}
			result.Matches[expr] = append(result.Matches[expr], matches...)
		}
	}

	return result, nil
}

func evaluateExpression(value cue.Value, expr string, file string, filters map[string]any) ([]Match, error) {
	path := cue.ParsePath(expr)
	if path.Err() != nil {
		return nil, path.Err()
	}

	matches := []Match{}
	matchedValue := value.LookupPath(path)
	if !matchedValue.Exists() {
		return matches, nil
	}

	// Handle different value types
	switch matchedValue.Kind() {
	case cue.StructKind:
		iter, _ := matchedValue.Fields()
		for iter.Next() {
			if match := extractMatch(iter.Value(), iter.Label(), file, filters); match != nil {
				matches = append(matches, *match)
			}
		}
	default:
		if match := extractMatch(matchedValue, path.String(), file, filters); match != nil {
			matches = append(matches, *match)
		}
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
		match.Value = path
		// Recursively process children
		iter, _ := value.Fields()
		for iter.Next() {
			if childMatch := extractMatch(iter.Value(), iter.Label(), file, filters); childMatch != nil {
				match.Children = append(match.Children, *childMatch)
			}
		}
	}

	return match
}

func FormatQueryResults(result QueryResult) string {
	var output strings.Builder

	if len(result.Matches) == 0 {
		return "No matches found in the configurations.\n"
	}

	for expr, matches := range result.Matches {
		fmt.Fprintf(&output, "Expression: %s\n", expr)
		for _, match := range matches {
			formatMatch(&output, match, 1)
		}
		output.WriteString("\n")
	}

	return output.String()
}

func formatMatch(output *strings.Builder, match Match, indent int) {
	indentStr := strings.Repeat("  ", indent)
	fmt.Fprintf(output, "%sPath: %s\n", indentStr, match.Path)
	fmt.Fprintf(output, "%sValue: %s\n", indentStr, match.Value)
	fmt.Fprintf(output, "%sType: %s\n", indentStr, match.Type)
	fmt.Fprintf(output, "%sFile: %s\n", indentStr, match.File)

	if len(match.Children) > 0 {
		fmt.Fprintf(output, "%sChildren:\n", indentStr)
		for _, child := range match.Children {
			formatMatch(output, child, indent+1)
		}
	}
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

// Add a helper function to check if a string is a regex pattern
func isRegexPattern(s string) bool {
	return strings.HasPrefix(s, "^") || strings.HasSuffix(s, "$") ||
		strings.Contains(s, "*") || strings.Contains(s, ".*")
}

func isMatchingValue(value cue.Value, filters map[string]any) bool {
	// Check type filter if present
	if typeFilter, ok := filters["type"]; ok {
		valueType := getValueType(value)
		switch tf := typeFilter.(type) {
		case string:
			if valueType != tf {
				return false
			}
		case []string:
			matched := false
			for _, t := range tf {
				if valueType == t {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
	}

	// Check other filters
	for key, filterValue := range filters {
		if key == "type" {
			continue // Already handled
		}

		// Handle nested filter paths (e.g., "env.REDIS_URL")
		path := cue.ParsePath(key)
		if path.Err() != nil {
			continue
		}

		matchedValue := value.LookupPath(path)
		if !matchedValue.Exists() {
			return false
		}

		// Convert filter value to string for comparison
		filterStr, ok := filterValue.(string)
		if !ok {
			continue
		}

		// Get the actual value as string
		str, err := matchedValue.String()
		if err != nil {
			continue
		}

		// Check if it's a regex pattern
		if isRegexPattern(filterStr) {
			re, err := regexp.Compile(filterStr)
			if err != nil {
				continue // Invalid regex pattern
			}
			if !re.MatchString(str) {
				return false
			}
		} else {
			// Exact string match
			if str != filterStr {
				return false
			}
		}
	}

	return true
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
