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
	"time"

	"cuelang.org/go/cue/cuecontext"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
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

	systemPrompt, err := loadPromptFromPath(systemPromptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load system prompt: %w", err)
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

	var queryConfig *mantis.QueryConfig
	var err error

	// Handle natural language query if provided
	if q.UserPrompt != "" {
		queryConfig, err = q.convertNaturalLanguageToQuery()
		if err != nil {
			return fmt.Errorf("failed to convert natural language query: %w", err)
		}
	} else {
		// Load query from config file
		var config mantis.QueryConfig
		config, err = mantis.LoadQueryConfig(q.QueryConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load query configuration: %w", err)
		}
		queryConfig = &config
	}

	fmt.Printf("=== Query Config ===\n%v\n==================\n", *queryConfig)

	results, err := mantis.QueryConfigurations(q.CodeDir, *queryConfig)
	if err != nil {
		return fmt.Errorf("failed to query configurations: %w", err)
	}

	formattedResults := mantis.FormatQueryResults(results, *queryConfig)
	if len(formattedResults) == 0 {
		return fmt.Errorf("no results found matching the query")
	}

	if len(formattedResults) > q.MaxResultSize {
		return fmt.Errorf("results exceed maximum size limit of %d bytes", q.MaxResultSize)
	}

	fmt.Println(formattedResults)
	return nil
}

func (q *Query) convertNaturalLanguageToQuery() (*mantis.QueryConfig, error) {
	// Load the index file
	indexData, err := os.ReadFile(q.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var sampleQueries []mantis.SampleQuery
	if err := json.Unmarshal(indexData, &sampleQueries); err != nil {
		return nil, fmt.Errorf("failed to parse index file: %w", err)
	}

	// Use AI to find the most relevant query from the index
	ctx := context.Background()

	combinedPrompt := fmt.Sprintf("System: %s\n\nUser: %s\n\nGiven these sample queries:\n%s\n\nPlease select and adapt the most relevant query to match the user's intent.",
		q.SystemPrompt,
		q.UserPrompt,
		string(indexData))

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

	fmt.Printf("=== JSON Output ===\n%s\n=================\n", string(jsonBytes))

	var queryConfig mantis.QueryConfig
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
	if q.UserPrompt != "" && q.IndexPath == "" {
		return fmt.Errorf("index path is required when using natural language query")
	}
	if _, err := os.Stat(q.CodeDir); err != nil {
		return fmt.Errorf("invalid code directory: %w", err)
	}
	return nil
}
