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
	"path/filepath"
	"time"

	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
)

type Index struct {
	AIGen        *codegen.AiGen
	SystemPrompt string
	CodeDir      string
	CacheDir     string
	Timeout      time.Duration
}

func NewIndex(confPath, systemPromptPath, codeDir, cacheDir string) (*Index, error) {
	aigen, err := codegen.New(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed initializing Index: %w", err)
	}

	systemPrompt, err := loadPromptFromPath(systemPromptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load system prompt: %w", err)
	}

	return &Index{
		AIGen:        aigen,
		SystemPrompt: systemPrompt,
		CodeDir:      codeDir,
		CacheDir:     cacheDir,
	}, nil
}

func (i *Index) Run() error {
	queries, err := i.generateSampleQueries()
	if err != nil {
		return fmt.Errorf("failed to generate sample queries: %w", err)
	}

	indexPath := filepath.Join(i.CacheDir, "mantis-index.json")
	if err := mantis.SaveQueries(indexPath, queries); err != nil {
		return fmt.Errorf("failed to save queries: %w", err)
	}

	fmt.Printf("Successfully indexed new queries to %s\n", indexPath)
	return nil
}

func (i *Index) generateSampleQueries() ([]mantis.SampleQuery, error) {
	ctx := context.Background()

	chat, err := i.AIGen.Chat(ctx, "", "")
	if err != nil {
		return nil, err
	}

	configs, err := mantis.LoadAllConfigurations(i.CodeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configurations: %w", err)
	}

	prompt := fmt.Sprintf(`%s
Analyze these CUE configurations and generate a list of important questions users might ask:
1. Write it in natural language
2. Provide a Mantis Query Language query to answer it
3. Explain why this question is important %s`, i.SystemPrompt, configs)

	response, err := chat.Send(ctx, prompt)
	if err != nil {
		return nil, err
	}

	fmt.Println(response.FullOutput)

	return mantis.ParseSampleQueries(response.FullOutput)
}
