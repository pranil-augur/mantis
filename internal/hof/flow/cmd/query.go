/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen/types"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
)

type Query struct {
	AIGen           *codegen.AiGen
	SystemPrompt    string
	UserPrompt      string
	CodeDir         string
	QueryConfigPath string
	MaxResultSize   int
	Timeout         time.Duration
}

func NewQuery(confPath, systemPromptPath, codeDir, userPrompt, queryConfigPath string) (*Query, error) {
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
		MaxResultSize:   10240,
		Timeout:         10 * time.Second,
	}, nil
}

func (q *Query) Run() error {
	// ctx, cancel := context.WithTimeout(context.Background(), q.Timeout)
	// defer cancel()

	if err := q.validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("Starting query with task: %s\n", q.UserPrompt)
	fmt.Printf("Using code directory: %s\n", q.CodeDir)
	fmt.Printf("Using query config: %s\n", q.QueryConfigPath)

	// chat, err := q.AIGen.Chat(ctx, "", "")
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize chat: %w", err)
	// }

	queryConfig, err := mantis.LoadQueryConfig(q.QueryConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load query configuration: %w", err)
	}
	fmt.Printf("Loaded query config: %+v\n", queryConfig)

	results, err := mantis.QueryConfigurations(q.CodeDir, queryConfig)
	if err != nil {
		return fmt.Errorf("failed to query configurations: %w", err)
	}
	// fmt.Printf("Raw query results: %+v\n", results)

	formattedResults := mantis.FormatQueryResults(results)
	fmt.Printf("Formatted results length: %d\n", len(formattedResults))

	if len(formattedResults) == 0 {
		return fmt.Errorf("no results found matching the query")
	}

	if len(formattedResults) > q.MaxResultSize {
		return fmt.Errorf("results exceed maximum size limit of %d bytes", q.MaxResultSize)
	}

	// response, err := q.generateResponse(ctx, chat, formattedResults)
	// if err != nil {
	// 	return fmt.Errorf("failed to generate response: %w", err)
	// }

	fmt.Println("Query response:")
	fmt.Println(formattedResults)

	return nil
}

func (q *Query) validate() error {
	if q.CodeDir == "" {
		return fmt.Errorf("code directory is required")
	}
	if q.QueryConfigPath == "" {
		return fmt.Errorf("query config path is required")
	}
	if _, err := os.Stat(q.CodeDir); err != nil {
		return fmt.Errorf("invalid code directory: %w", err)
	}
	if _, err := os.Stat(q.QueryConfigPath); err != nil {
		return fmt.Errorf("invalid query config path: %w", err)
	}
	return nil
}

func (q *Query) generateResponse(ctx context.Context, chat types.Conversation, queryResults string) (string, error) {
	prompt := &PromptTemplate{
		System:       q.SystemPrompt,
		User:         q.UserPrompt,
		QueryResults: queryResults,
	}

	combinedPrompt, err := prompt.Format()
	if err != nil {
		return "", fmt.Errorf("failed to format prompt: %w", err)
	}

	response, err := chat.Send(ctx, combinedPrompt)
	if err != nil {
		return "", err
	}

	return response.FullOutput, nil
}

type PromptTemplate struct {
	System       string
	User         string
	QueryResults string
}

func (p *PromptTemplate) Format() (string, error) {
	if err := p.validate(); err != nil {
		return "", err
	}

	return fmt.Sprintf(`System: %s

User Question: %s

Query Results:
%s

Please analyze these results and provide insights:
`, p.System, p.User, p.QueryResults), nil
}

func (p *PromptTemplate) validate() error {
	if p.System == "" {
		return fmt.Errorf("system prompt is required")
	}
	if p.User == "" {
		return fmt.Errorf("user prompt is required")
	}
	if p.QueryResults == "" {
		return fmt.Errorf("query results are required")
	}
	return nil
}
