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
	"path/filepath"
	"time"

	"github.com/opentofu/opentofu/internal/hof/lib/codegen"
	cql "github.com/opentofu/opentofu/internal/hof/lib/mantis/cql"
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
	indexPath := filepath.Join(i.CacheDir, "mantis-index.json")

	fmt.Println("Building Mantis index...")
	if err := cql.BuildIndex(i.CodeDir, indexPath); err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// printIndex(indexPath)
	// Generate sample queries
	queries, err := i.generateSampleQueries()
	if err != nil {
		return fmt.Errorf("failed to generate sample queries: %w", err)
	}

	// Load existing index
	metadata, err := cql.LoadIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	// Update index with sample queries
	metadata.SampleQueries = queries

	// Save updated index
	if err := cql.SaveIndex(indexPath, metadata); err != nil {
		return fmt.Errorf("failed to save updated index: %w", err)
	}

	fmt.Printf("Successfully updated index with queries at %s\n", indexPath)
	return nil
}

func (i *Index) generateSampleQueries() ([]cql.SampleQuery, error) {
	ctx := context.Background()

	chat, err := i.AIGen.Chat(ctx, "", "")
	if err != nil {
		return nil, err
	}

	configs, err := cql.LoadAllConfigurations(i.CodeDir)
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

	return cql.ParseSampleQueries(response.FullOutput)
}

func printIndex(indexPath string) error {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
