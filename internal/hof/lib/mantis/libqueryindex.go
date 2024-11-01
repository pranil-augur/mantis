/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package mantis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
)

type IndexMetadata struct {
	SampleQueries []SampleQuery `json:"sample_queries"`
	ConfigPaths   []string      `json:"config_paths"` // List of all configuration paths
	Types         []string      `json:"types"`        // List of discovered types (service, resource, etc)
}

type SampleQuery struct {
	NaturalLanguage string      `json:"natural_language"` // The question in natural language
	MantisQuery     QueryConfig `json:"mantis_query"`     // The corresponding Mantis query
	Description     string      `json:"description"`      // Why this query is useful
}

// BuildIndex analyzes configurations and generates sample queries
func BuildIndex(directory string, aiGen *codegen.AiGen) error {
	// Load all configurations
	configs, err := LoadAllConfigurations(directory)
	if err != nil {
		return fmt.Errorf("failed to load configurations: %w", err)
	}

	// Generate sample queries using AI
	queries, err := generateSampleQueries(configs, aiGen)
	if err != nil {
		return fmt.Errorf("failed to generate sample queries: %w", err)
	}

	// Create index metadata
	metadata := IndexMetadata{
		SampleQueries: queries,
		ConfigPaths:   extractConfigPaths(configs),
		Types:         extractConfigTypes(configs),
	}

	// Save the index
	indexPath := filepath.Join(directory, ".mantis-index.json")
	indexData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return os.WriteFile(indexPath, indexData, 0644)
}

func generateSampleQueries(configs string, aiGen *codegen.AiGen) ([]SampleQuery, error) {
	ctx := context.Background()

	// Initialize chat with system prompt
	chat, err := aiGen.Chat(ctx, `You are a CUE configuration analyzer. Given a set of CUE configurations:
1. Identify important questions users might ask about these configurations
2. Generate corresponding Mantis Query Language queries to answer these questions
3. Focus on operational, security, and dependency-related questions`, "")

	if err != nil {
		return nil, err
	}

	// Ask AI to analyze configs and generate questions
	prompt := fmt.Sprintf(`Analyze these CUE configurations and generate a list of important questions users might ask:

%s

For each question:
1. Write it in natural language
2. Provide a Mantis Query Language query to answer it
3. Explain why this question is important

Use the query format from the Mantis Query Language specification:
query: {
    from: string      // Path-based data source
    select: [...string] // Fields to retrieve
    where: [string]: string // Filtering conditions
}`, configs)

	response, err := chat.Send(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse the response into sample queries
	queries, err := ParseSampleQueries(response.FullOutput)
	if err != nil {
		return nil, err
	}

	return queries, nil
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
