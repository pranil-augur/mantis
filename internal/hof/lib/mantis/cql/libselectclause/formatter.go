package libselectclause

import (
	"fmt"
	"path/filepath"
	"strings"

	types "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/shared"
)

// FormatResults formats the query results into a table-like string
func FormatResults(result types.QueryResult, config types.QueryConfig) string {
	var output strings.Builder

	if len(result.Matches) == 0 {
		return "No matches found in the configurations.\n"
	}

	// Determine fields to display
	fields := determineDisplayFields(result, config.Select)

	// Print header
	printHeader(&output, fields)

	// Print separator
	printSeparator(&output, fields)

	// Print values
	for _, matches := range result.Matches {
		for _, match := range matches {
			FormatMatchAsTable(&output, match, fields)
		}
	}

	return output.String()
}

// FormatMatchAsTable formats a single match as a table row
func FormatMatchAsTable(output *strings.Builder, match types.Match, fields []string) {
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

func determineDisplayFields(result types.QueryResult, selects []string) []string {
	if len(selects) == 1 && selects[0] == "*" {
		fieldSet := make(map[string]bool)
		for _, matches := range result.Matches {
			for _, match := range matches {
				fieldSet[match.Path] = true
				for _, child := range match.Children {
					fieldSet[match.Path+"."+child.Path] = true
				}
			}
		}
		var fields []string
		for field := range fieldSet {
			fields = append(fields, field)
		}
		return fields
	}
	return selects
}

func printHeader(output *strings.Builder, fields []string) {
	fmt.Fprintf(output, "%-30s", "file")
	for _, h := range fields {
		fmt.Fprintf(output, "%-20s", h)
	}
	output.WriteString("\n")
}

func printSeparator(output *strings.Builder, fields []string) {
	output.WriteString(strings.Repeat("-", 30))
	for range fields {
		output.WriteString(strings.Repeat("-", 20))
	}
	output.WriteString("\n")
}

func findFieldValue(match types.Match, field string) string {
	// Handle wildcard selector
	if field == "*" {
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
