/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type QueryConfig struct {
	TargetFields []string          `json:"targetFields"`
	TargetTypes  []string          `json:"targetTypes"`
	Filters      map[string]string `json:"filters"`
}

type QueryResult struct {
	MatchedFields map[string][]string
}

func LoadQueryConfig(path string) (QueryConfig, error) {
	var config QueryConfig
	file, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func QueryConfigurations(directory string, config QueryConfig) (QueryResult, error) {
	files, err := getCueFiles(directory)
	if err != nil {
		return QueryResult{}, err
	}

	result := QueryResult{MatchedFields: make(map[string][]string)}
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

		searchValue(value, file, config, &result)
	}

	return result, nil
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

func searchValue(value cue.Value, file string, config QueryConfig, result *QueryResult) {
	iter, _ := value.Fields()
	for iter.Next() {
		field := iter.Label()
		fieldValue := iter.Value()

		if isTargetField(field, fieldValue, config) && matchesFilters(fieldValue, config.Filters) {
			matchedValue := getMatchedValue(fieldValue)
			if matchedValue != "" {
				result.MatchedFields[matchedValue] = append(result.MatchedFields[matchedValue], file)
			}
		}

		// Recursively search nested structures
		if fieldValue.Kind() == cue.StructKind {
			searchValue(fieldValue, file, config, result)
		}
	}
}

func isTargetField(field string, value cue.Value, config QueryConfig) bool {
	// Check for specific field names
	for _, targetField := range config.TargetFields {
		if strings.Contains(field, targetField) {
			return true
		}
	}

	// Check for specific types (Terraform resources or K8s kinds)
	for _, targetType := range config.TargetTypes {
		// Check for Terraform-style "type" field
		if value.LookupPath(cue.ParsePath("type")).Exists() {
			typeValue, err := value.LookupPath(cue.ParsePath("type")).String()
			if err == nil && typeValue == targetType {
				return true
			}
		}
		// Check for Kubernetes-style "kind" field
		if value.LookupPath(cue.ParsePath("kind")).Exists() {
			kindValue, err := value.LookupPath(cue.ParsePath("kind")).String()
			if err == nil && kindValue == targetType {
				return true
			}
		}
	}

	return false
}

func getMatchedValue(value cue.Value) string {
	// Try to get the "type" field (Terraform)
	typeValue := value.LookupPath(cue.ParsePath("type"))
	if typeValue.Exists() {
		if typeStr, err := typeValue.String(); err == nil {
			return typeStr
		}
	}

	// Try to get the "kind" field (Kubernetes)
	kindValue := value.LookupPath(cue.ParsePath("kind"))
	if kindValue.Exists() {
		if kindStr, err := kindValue.String(); err == nil {
			return kindStr
		}
	}

	// For other elements, try to get the name
	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		return name
	}

	// If all else fails, return an empty string
	return ""
}

func FormatQueryResults(result QueryResult) string {
	var output strings.Builder

	if len(result.MatchedFields) == 0 {
		output.WriteString("No matches found in the configurations.\n")
		return output.String()
	}

	output.WriteString("Matches found in configurations:\n")
	for matchedValue, files := range result.MatchedFields {
		fmt.Fprintf(&output, "Matched: %s\n", matchedValue)
		output.WriteString("Found in files:\n")
		for _, file := range files {
			fmt.Fprintf(&output, "  - %s\n", file)
		}
		output.WriteString("\n")
	}

	return output.String()
}

func matchesFilters(value cue.Value, filters map[string]string) bool {
	for key, expectedValue := range filters {
		fieldValue := value.LookupPath(cue.ParsePath(key))
		if !fieldValue.Exists() {
			return false
		}
		actualValue, err := fieldValue.String()
		if err != nil || actualValue != expectedValue {
			return false
		}
	}
	return true
}
