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
	}, nil
}

func (q *Query) Run() error {
	ctx := context.Background()
	fmt.Printf("Starting query with task: %s\n", q.UserPrompt)

	chat, err := q.AIGen.Chat(ctx, "", "")
	if err != nil {
		return fmt.Errorf("failed to initialize chat: %w", err)
	}

	queryConfig, err := mantis.LoadQueryConfig(q.QueryConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load query configuration: %w", err)
	}

	results, err := mantis.QueryConfigurations(q.CodeDir, queryConfig)
	if err != nil {
		return fmt.Errorf("failed to query configurations: %w", err)
	}

	formattedResults := mantis.FormatQueryResults(results)

	response, err := q.generateResponse(ctx, chat, formattedResults)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	fmt.Println("Query response:")
	fmt.Println(response)

	return nil
}

func (q *Query) generateResponse(ctx context.Context, chat types.Conversation, queryResults string) (string, error) {
	combinedPrompt := fmt.Sprintf("System: %s\n\nUser: %s\n\nGiven the following query results, please analyze and provide insights:\n\nQuery Results:\n%s\n\nAnalysis:\n",
		q.SystemPrompt, q.UserPrompt, queryResults)

	response, err := chat.Send(ctx, combinedPrompt)
	if err != nil {
		return "", err
	}

	return response.FullOutput, nil
}
