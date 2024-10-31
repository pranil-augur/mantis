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
	Select []string       `json:"select"` // CUE path expressions for selection
	Where  map[string]any `json:"where"`  // Predicate conditions
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

	// Extract select expressions
	if selects, err := extractStringSlice(value, "select"); err == nil {
		config.Select = selects
	}

	// Extract where predicates
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

		// Process each expression
		for _, expr := range config.Select {
			matches, err := evaluateExpression(value, expr, file, config.Where)
			if err != nil {
				continue
			}
			result.Matches[expr] = append(result.Matches[expr], matches...)
		}
	}

	return result, nil
}

func evaluateExpression(value cue.Value, expr string, file string, filters map[string]any) ([]Match, error) {
	// Check if it's a pattern expression
	if prefix, pattern, suffix, ok := parsePatternExpression(expr); ok {
		// Get the root value
		rootPath := cue.ParsePath(prefix)
		if rootPath.Err() != nil {
			return nil, rootPath.Err()
		}

		rootValue := value.LookupPath(rootPath)
		if !rootValue.Exists() {
			return nil, nil
		}

		matches := []Match{}

		// Iterate over fields matching the pattern
		iter, _ := rootValue.Fields()
		for iter.Next() {
			fieldValue := iter.Value()

			// If there's a suffix, look it up
			if suffix != "" {
				fieldValue = fieldValue.LookupPath(cue.ParsePath(suffix))
				if !fieldValue.Exists() {
					continue
				}
			}

			// Check if the value matches the pattern type
			if isMatchingPatternType(fieldValue, pattern) {
				if match := extractMatch(fieldValue, iter.Label(), file, filters); match != nil {
					matches = append(matches, *match)
				}
			}
		}

		return matches, nil
	}

	// Handle non-pattern expressions (existing code)
	path := cue.ParsePath(expr)
	if path.Err() != nil {
		return nil, path.Err()
	}

	matches := []Match{}
	matchedValue := value.LookupPath(path)
	if !matchedValue.Exists() {
		return matches, nil
	}

	if match := extractMatch(matchedValue, path.String(), file, filters); match != nil {
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
	for filterPath, filterValue := range filters {
		filterStr, ok := filterValue.(string)
		if !ok {
			continue
		}

		// Get the value at the filter path
		matchedValue := value.LookupPath(cue.ParsePath(filterPath))
		if !matchedValue.Exists() {
			return false
		}

		// Get the actual value as string
		valueStr, err := matchedValue.String()
		if err != nil {
			return false
		}

		// Try regex first, fall back to exact match
		if re, err := regexp.Compile(filterStr); err == nil {
			if !re.MatchString(valueStr) {
				return false
			}
		} else if valueStr != filterStr {
			// If not a valid regex, do exact match
			return false
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
		return value.Kind() == cue.StringKind
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
