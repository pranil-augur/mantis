/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 */

package cql

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	fromClause "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/libfromclause"
	selectClause "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/libselectclause"
	whereClause "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/libwhereclause"
	types "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/shared"
)

// LoadQueryConfig loads and parses a CUE query file
func LoadQueryConfig(path string) (types.QueryConfig, error) {
	var config types.QueryConfig

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
		fmt.Printf("WHERE clause exists: %v\n", where)

		iter, err := where.Fields()
		if err != nil {
			fmt.Printf("Error getting fields: %v\n", err)
			return config, fmt.Errorf("failed to get where clause fields: %v", err)
		}

		var rawWhere interface{}
		if err := where.Decode(&rawWhere); err == nil {
			fmt.Printf("Raw WHERE value: %+v\n", rawWhere)
		}

		for iter.Next() {
			fmt.Printf("Processing field: %s\n", iter.Label())
			// Try to get the value in its native form instead of forcing string conversion
			val := iter.Value()
			if val.Exists() {
				// For debugging
				fmt.Printf("WHERE clause key: %s, value: %v\n", iter.Label(), val)

				// Try to decode the value into a generic interface{}
				var decoded interface{}
				if err := val.Decode(&decoded); err == nil {
					config.Where[iter.Label()] = decoded
				} else {
					// Fallback to string if decode fails
					if str, err := val.String(); err == nil {
						config.Where[iter.Label()] = str
					}
				}
			}
		}

		if err := where.Decode(&config.Where); err != nil {
			fmt.Printf("Failed to decode WHERE directly: %v\n", err)
		}
	}

	return config, nil
}

// QueryConfigurations executes the query against CUE files in the specified directory
func QueryConfigurations(directory string, config types.QueryConfig) (types.QueryResult, error) {
	result := types.QueryResult{
		Matches: make(map[string][]types.Match),
	}

	// Create a single accumulator for all matches
	accumulatedMatch := types.Match{
		Fields: make(map[string]interface{}),
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

		// Process FROM clause
		if config.From != "" {
			if prefix, pattern, suffix, ok := fromClause.ParsePatternExpression(config.From); ok {
				// Handle pattern-based FROM
				matches, err := fromClause.EvaluatePattern(value, prefix, pattern, suffix)
				if err != nil {
					continue
				}

				// Process matches with WHERE and SELECT
				for _, match := range matches {
					if len(config.Where) > 0 {
						whereValue := ctx.Encode(config.Where)
						evaluator := whereClause.CreateEvaluator(whereValue)
						if !evaluator.Evaluate(match.CueValue) {
							continue
						}
					}

					// Store the file path with underscore prefix
					accumulatedMatch.Fields["_file"] = file

					// Accumulate fields into the single match
					for _, proj := range config.Select {
						if proj == "_file" {
							continue // Skip as we've already handled it
						}
						if val := match.CueValue.LookupPath(cue.ParsePath(proj)); val.Exists() {
							var decoded interface{}
							if err := val.Decode(&decoded); err == nil {
								// For arrays/slices, append to existing values
								if existing, ok := accumulatedMatch.Fields[proj]; ok {
									if existingSlice, ok := existing.([]interface{}); ok {
										if newSlice, ok := decoded.([]interface{}); ok {
											accumulatedMatch.Fields[proj] = append(existingSlice, newSlice...)
											continue
										}
									}
								}
								accumulatedMatch.Fields[proj] = decoded
							}
						}
					}
				}
			} else {
				// Handle direct path FROM
				baseValue := value.LookupPath(cue.ParsePath(config.From))
				if !baseValue.Exists() {
					continue
				}

				// Apply WHERE and SELECT
				if len(config.Where) > 0 {
					whereValue := ctx.Encode(config.Where)
					evaluator := whereClause.CreateEvaluator(whereValue)
					if !evaluator.Evaluate(baseValue) {
						continue
					}
				}
				// Accumulate fields into the single match
				for _, proj := range config.Select {
					if val := baseValue.LookupPath(cue.ParsePath(proj)); val.Exists() {
						var decoded interface{}
						if err := val.Decode(&decoded); err == nil {
							// For arrays/slices, append to existing values
							if existing, ok := accumulatedMatch.Fields[proj]; ok {
								if existingSlice, ok := existing.([]interface{}); ok {
									if newSlice, ok := decoded.([]interface{}); ok {
										accumulatedMatch.Fields[proj] = append(existingSlice, newSlice...)
										continue
									}
								}
							}
							accumulatedMatch.Fields[proj] = decoded
						}
					}
				}
			}
		}
	}

	// Add the accumulated match to the result only once
	if len(accumulatedMatch.Fields) > 0 {
		result.Matches["accumulated"] = []types.Match{accumulatedMatch}
	}

	return result, nil
}

// FormatQueryResults formats the query results as a table
func FormatQueryResults(result types.QueryResult, config types.QueryConfig) string {
	return selectClause.FormatResults(result, config)
}

// Helper function to get CUE files from directory
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

// Helper function to extract string slice from CUE value
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
