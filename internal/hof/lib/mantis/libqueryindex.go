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
	"math"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

type SchemaIndex struct {
	Fields map[string]string `json:"fields"` // Maps field paths to their types
}

type IndexMetadata struct {
	SampleQueries []SampleQuery `json:"sample_queries"`
	ConfigPaths   []string      `json:"config_paths"`
	Types         []string      `json:"types"`
	Values        ValueIndex    `json:"values"`
	Schema        SchemaIndex   `json:"schema"`
}

type ValueIndex struct {
	NumericFields  map[string]NumericFieldInfo `json:"numeric_fields"`
	StringFields   map[string]StringFieldInfo  `json:"string_fields"`
	ComputedValues map[string]float64          `json:"computed_values"`
	Aggregations   map[string]interface{}      `json:"aggregations"`
}

type NumericFieldInfo struct {
	PathPattern string             `json:"path_pattern"`
	Type        string             `json:"type"`
	Occurrences map[string]float64 `json:"occurrences"`
	Stats       NumericStats       `json:"stats"`
}

type NumericStats struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Total   float64 `json:"total"`
	Average float64 `json:"average"`
}

type StringFieldInfo struct {
	PathPattern  string         `json:"path_pattern"`
	UniqueValues []string       `json:"unique_values"`
	Occurrences  map[string]int `json:"occurrences"`
}

type SampleQuery struct {
	NaturalLanguage string      `json:"natural_language"` // The question in natural language
	MantisQuery     QueryConfig `json:"mantis_query"`     // The corresponding Mantis query
	Description     string      `json:"description"`      // Why this query is useful
}

// BuildIndex analyzes configurations and generates sample queries
func BuildIndex(directory string, indexPath string) error {
	fmt.Printf("Schema: Starting build index for directory %s\n", directory)

	// Initialize indexes
	schemaIndex := SchemaIndex{
		Fields: make(map[string]string),
	}
	valueIndex := ValueIndex{
		NumericFields:  make(map[string]NumericFieldInfo),
		StringFields:   make(map[string]StringFieldInfo),
		ComputedValues: make(map[string]float64),
		Aggregations:   make(map[string]interface{}),
	}

	// Load CUE files directly
	files, err := filepath.Glob(filepath.Join(directory, "*.cue"))
	if err != nil {
		return fmt.Errorf("failed to find CUE files: %w", err)
	}

	ctx := cuecontext.New()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("DEBUG: Error reading file %s: %v\n", file, err)
			continue
		}

		value := ctx.CompileString(string(data))
		if value.Err() != nil {
			fmt.Printf("DEBUG: Error compiling CUE from %s: %v\n", file, value.Err())
			continue
		}

		if err := walkAndCollectSchema(value, "", &schemaIndex); err != nil {
			fmt.Printf("DEBUG: Error collecting schema: %v\n", err)
			return err
		}
		if err := walkAndCollectValues(value, "", &valueIndex); err != nil {
			fmt.Printf("DEBUG: Error collecting values: %v\n", err)
			return err
		}
	}

	// fmt.Printf("DEBUG: Final schema fields: %+v\n", schemaIndex.Fields)
	// fmt.Printf("DEBUG: Final value fields: %+v\n", valueIndex)

	// Create and save the index
	metadata := IndexMetadata{
		ConfigPaths: []string{directory},
		Types:       []string{},
		Values:      valueIndex,
		Schema:      schemaIndex,
	}

	// Save to index file
	if err := SaveIndex(indexPath, metadata); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// Add this new function
func SaveIndex(path string, metadata IndexMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// ParseSampleQueries converts an AI-generated response into structured SampleQuery objects.
// The function expects the response to contain sections for each query, with sections
// separated by double newlines (\n\n). Each query section should follow this format:
//
// Question: What services are exposed to the internet?
// or
// 1. What services are exposed to the internet?
//
//	query: {
//	    "from": "services",
//	    "select": ["name", "ports"],
//	    "where": {
//	        "exposed": "true"
//	    }
//	}
//
// Why: This helps identify potential security risks by showing externally accessible services.
// or
// 3. This helps identify potential security risks by showing externally accessible services.
//
// The parser recognizes:
// - Questions prefixed with "Question:" or numbered "1."
// - Query configurations in JSON format starting with "query:"
// - Descriptions prefixed with "Why:", "Important:", or numbered "3."
//
// Returns a slice of SampleQuery objects or an error if parsing fails.
func ParseSampleQueries(response string) ([]SampleQuery, error) {
	if response == "" {
		return nil, fmt.Errorf("empty response from AI")
	}

	ctx := cuecontext.New()
	queryBlocks := strings.Split(response, "---")
	var queries []SampleQuery

	for _, block := range queryBlocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		// Parse CUE to JSON
		value := ctx.CompileString(block)
		if value.Err() != nil {
			fmt.Printf("Error parsing CUE: %v\n", value.Err())
			continue
		}

		fmt.Printf("CUE value: %v\n", value)

		jsonBytes, err := value.MarshalJSON()
		if err != nil {
			fmt.Printf("Error converting to JSON: %v\n", err)
			continue
		}

		var queryConfig QueryConfig
		if err := json.Unmarshal(jsonBytes, &queryConfig); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			continue
		}

		query := SampleQuery{
			MantisQuery:     queryConfig,
			NaturalLanguage: fmt.Sprintf("Query for %s", queryConfig.From),
			Description:     "Auto-generated query",
		}
		queries = append(queries, query)
	}

	if len(queries) == 0 {
		return nil, fmt.Errorf("no valid queries found in response")
	}

	return queries, nil
}

// extractContent extracts the meaningful content from a text section by removing
// prefixes (like "Question:", "Why:", etc.) or numerical markers.
// If a colon is found, returns everything after it; otherwise returns everything
// after the first word.
func extractContent(text string) string {
	parts := strings.SplitN(text, ":", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	// If no colon found, return everything after the first word
	words := strings.Fields(text)
	if len(words) > 1 {
		return strings.Join(words[1:], " ")
	}
	return text
}

// extractBetween returns the substring between the first occurrence of 'start'
// and the last occurrence of 'end' in the given text.
// Returns an empty string if either delimiter is not found.
func extractBetween(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return ""
	}
	text = text[startIdx:]
	endIdx := strings.LastIndex(text, end)
	if endIdx == -1 {
		return ""
	}
	return text[:endIdx+1]
}

// LoadAllConfigurations reads all CUE configuration files from the given directory
func LoadAllConfigurations(directory string) (string, error) {
	files, err := filepath.Glob(filepath.Join(directory, "*.cue"))
	if err != nil {
		return "", fmt.Errorf("failed to glob cue files: %w", err)
	}

	var allConfigs string
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", file, err)
		}
		allConfigs += string(content) + "\n"
	}
	return allConfigs, nil
}

// extractConfigPaths returns a list of all configuration file paths
func extractConfigPaths(configs string) []string {
	paths, _ := filepath.Glob(filepath.Join(".", "*.cue"))
	return paths
}

// extractConfigTypes extracts unique configuration types from the configs
func extractConfigTypes(configs string) []string {
	// For now, return empty slice - implement type extraction logic later
	return []string{}
}

// SaveQueries writes the sample queries to the specified file path
func SaveQueries(path string, queries []SampleQuery) error {
	data, err := json.MarshalIndent(queries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal queries: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func walkAndCollectValues(v cue.Value, path string, index *ValueIndex) error {
	switch v.Kind() {
	case cue.IntKind, cue.FloatKind:
		var num float64
		if err := v.Decode(&num); err == nil {
			updateNumericStats(path, num, index)
		}

	case cue.StringKind:
		var str string
		if err := v.Decode(&str); err == nil {
			updateStringStats(path, str, index)
		}

	case cue.StructKind:
		iter, _ := v.Fields()
		for iter.Next() {
			newPath := path
			if newPath == "" {
				newPath = iter.Label()
			} else {
				newPath = newPath + "." + iter.Label()
			}
			if err := walkAndCollectValues(iter.Value(), newPath, index); err != nil {
				return err
			}
		}
	}

	return nil
}

func updateNumericStats(path string, value float64, index *ValueIndex) {
	info := index.NumericFields[path]
	if info.Occurrences == nil {
		info.Occurrences = make(map[string]float64)
		info.Type = "number"
		info.PathPattern = path
	}

	info.Occurrences[path] = value
	info.Stats.Total += value
	info.Stats.Min = math.Min(info.Stats.Min, value)
	info.Stats.Max = math.Max(info.Stats.Max, value)

	count := float64(len(info.Occurrences))
	info.Stats.Average = info.Stats.Total / count

	index.NumericFields[path] = info

	// Update computed values for specific fields
	if strings.HasSuffix(path, ".replicas") {
		index.ComputedValues["total_replicas"] = info.Stats.Total
		index.ComputedValues["average_replicas"] = info.Stats.Average
	}
}

func updateStringStats(path string, value string, index *ValueIndex) {
	info := index.StringFields[path]
	if info.Occurrences == nil {
		info.Occurrences = make(map[string]int)
		info.PathPattern = path
	}

	info.Occurrences[value]++
	if !contains(info.UniqueValues, value) {
		info.UniqueValues = append(info.UniqueValues, value)
	}

	index.StringFields[path] = info
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func walkAndCollectSchema(v cue.Value, path string, schema *SchemaIndex) error {
	if path != "" {
		schema.Fields[path] = v.Kind().String()
	}

	switch v.Kind() {
	case cue.StructKind:
		iter, _ := v.Fields()
		for iter.Next() {
			newPath := path
			if newPath == "" {
				newPath = iter.Label()
			} else {
				newPath = newPath + "." + iter.Label()
			}
			if err := walkAndCollectSchema(iter.Value(), newPath, schema); err != nil {
				return err
			}
		}
	case cue.ListKind:
		iter, err := v.List()
		if err != nil {
			return err
		}
		for iter.Next() {
			if err := walkAndCollectSchema(iter.Value(), path, schema); err != nil {
				return err
			}
		}
	}
	return nil
}

func LoadIndex(path string) (IndexMetadata, error) {
	var metadata IndexMetadata

	data, err := os.ReadFile(path)
	if err != nil {
		return metadata, fmt.Errorf("failed to read index file: %w", err)
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return metadata, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	return metadata, nil
}
