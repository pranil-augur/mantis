/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"cuelang.org/go/cue/cuecontext"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
	cql "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql"
	types "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql/shared"
)

type Query struct {
	AIGen           *codegen.AiGen
	SystemPrompt    string
	UserPrompt      string
	CodeDir         string
	QueryConfigPath string
	IndexPath       string
	MaxResultSize   int
	Timeout         time.Duration
}

func NewQuery(confPath, systemPromptPath, codeDir, userPrompt, queryConfigPath, indexPath string) (*Query, error) {
	aigen, err := codegen.New(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed initializing Query: %w", err)
	}

	var systemPrompt string
	// Only load system prompt if we're doing natural language query
	if userPrompt != "" {
		systemPrompt, err = loadPromptFromPath(systemPromptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load system prompt: %w", err)
		}
	}

	return &Query{
		AIGen:           aigen,
		SystemPrompt:    systemPrompt,
		UserPrompt:      userPrompt,
		CodeDir:         codeDir,
		QueryConfigPath: queryConfigPath,
		IndexPath:       indexPath,
		MaxResultSize:   10240,
		Timeout:         10 * time.Second,
	}, nil
}

func (q *Query) Run() error {
	if err := q.validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	var queryConfig *types.QueryConfig
	var err error

	// Handle natural language query if provided
	if q.UserPrompt != "" {
		queryConfig, err = q.convertNaturalLanguageToQuery()
		if err != nil {
			return fmt.Errorf("failed to convert natural language query: %w", err)
		}
	} else {
		// Load query from config file
		var config types.QueryConfig
		config, err = cql.LoadQueryConfig(q.QueryConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load query configuration: %w", err)
		}
		queryConfig = &config
	}

	fmt.Print(formatQueryConfig(queryConfig))

	results, err := cql.QueryConfigurations(q.CodeDir, *queryConfig)
	if err != nil {
		return fmt.Errorf("failed to query configurations: %w", err)
	}

	formattedResults := cql.FormatQueryResults(results, *queryConfig)
	if len(formattedResults) == 0 {
		return fmt.Errorf("no results found matching the query")
	}

	if len(formattedResults) > q.MaxResultSize {
		return fmt.Errorf("results exceed maximum size limit of %d bytes", q.MaxResultSize)
	}

	fmt.Println(formattedResults)
	return nil
}

func (q *Query) convertNaturalLanguageToQuery() (*types.QueryConfig, error) {
	// Load the index using the existing function
	metadata, err := cql.LoadIndex(q.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	// Use AI to find the most relevant query from the index
	ctx := context.Background()

	// Create a more informative prompt with schema and sample queries
	combinedPrompt := fmt.Sprintf(`System: %s

User: %s

Available Schema:
%+v

Sample Queries:
%+v

Please select and adapt the most relevant query to match the user's intent.`,
		q.SystemPrompt,
		q.UserPrompt,
		metadata.Schema,
		metadata.SampleQueries)

	chat, err := q.AIGen.Chat(ctx, "", "")
	if err != nil {
		return nil, err
	}

	response, err := chat.Send(ctx, combinedPrompt)
	if err != nil {
		return nil, err
	}

	// Parse the AI response into a QueryConfig
	cueCtx := cuecontext.New()
	value := cueCtx.CompileString(response.FullOutput)
	if value.Err() != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", value.Err())
	}

	jsonBytes, err := value.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to convert query to JSON: %w", err)
	}

	var queryConfig types.QueryConfig
	if err := json.Unmarshal(jsonBytes, &queryConfig); err != nil {
		return nil, fmt.Errorf("failed to parse query config: %w", err)
	}

	return &queryConfig, nil
}

func (q *Query) validate() error {
	if q.CodeDir == "" {
		return fmt.Errorf("code directory is required")
	}
	if q.UserPrompt == "" && q.QueryConfigPath == "" {
		return fmt.Errorf("either user prompt or query config path is required")
	}
	// Only validate system prompt and index for natural language queries
	if q.UserPrompt != "" {
		if q.SystemPrompt == "" {
			return fmt.Errorf("system prompt is required when using natural language query")
		}
		if q.IndexPath == "" {
			return fmt.Errorf("index path is required when using natural language query")
		}
	}
	if _, err := os.Stat(q.CodeDir); err != nil {
		return fmt.Errorf("invalid code directory: %w", err)
	}
	return nil
}

func formatQueryConfig(config *types.QueryConfig) string {
	var output strings.Builder

	output.WriteString("=== Configuration Query ===\n")

	// FROM clause
	output.WriteString("FROM: ")
	output.WriteString(config.From)
	output.WriteString("\n")

	// SELECT clause
	output.WriteString("SELECT: ")
	output.WriteString(strings.Join(config.Select, ", "))
	output.WriteString("\n")

	// WHERE clause
	if len(config.Where) > 0 {
		output.WriteString("WHERE: ")
		conditions := make([]string, 0)
		for k, v := range config.Where {
			conditions = append(conditions, fmt.Sprintf("%s = %v", k, v))
		}
		output.WriteString(strings.Join(conditions, " AND "))
		output.WriteString("\n")
	}

	output.WriteString("==================\n")
	return output.String()
}
