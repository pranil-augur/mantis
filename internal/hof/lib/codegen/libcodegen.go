/*
/*
 * Copyright (c) 2024 Augur AI, Inc.
 * This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
 * If a copy of the MPL was not distributed with this file, you can obtain one at https://mozilla.org/MPL/2.0/.
 *
 
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/gofireflyio/aiac, licensed under the MIT License.
 */
package codegen

import (
	"context"
	"fmt"

	"github.com/opentofu/opentofu/internal/hof/lib/codegen/openai"
	"github.com/opentofu/opentofu/internal/hof/lib/codegen/types"
)

// Version contains aiac's version string
var Version = "development"

type AiGen struct {
	// Conf holds the configuration for aiac.
	Conf Config

	// Backends is a map from backend names to backend implementations.
	Backends map[string]types.Backend
}

// New constructs a new Aiac object with the path to a configuration file. If
// a configuration file is not provided, the default path will be checked based
// on the XDG specification. On Unix-like operating systems, this will be
// ~/.config/aiac/aiac.toml.
func New(configPath ...string) (*AiGen, error) {
	path := ""
	if len(configPath) > 0 {
		path = configPath[0]
	}

	conf, err := LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed loading configuration: %w", err)
	}

	return &AiGen{Conf: conf}, nil
}

// NewFromConf is the same as New, but receives a populated configuration object
// rather than a file path.
func NewFromConf(conf Config) *AiGen {
	return &AiGen{Conf: conf}
}

// ListModels returns a list of all the models supported by the selected
// backend, identified by its name. If backendName is an empty string, the
// default backend defined in the configuration file will be used, if any.
func (aigen *AiGen) ListModels(ctx context.Context, backendName string) (
	models []string,
	err error,
) {
	backend, _, err := aigen.loadBackend(ctx, backendName)
	if err != nil {
		return models, fmt.Errorf("failed loading backend: %w", err)
	}

	return backend.ListModels(ctx)
}

// Chat initiates a chat conversation with the provided chat model of the
// selected backend. Returns a Conversation object with which messages can be
// sent and received. If backendName is an empty string, the default backend
// defined in the configuration will be used, if any. If model is an empty
// string, the default model defined in the backend configuration will be used,
// if any. Users can also supply zero or more "previous messages" that may have
// been exchanged in the past. This practically allows "loading" previous
// conversations and continuing them.
func (aigen *AiGen) Chat(
	ctx context.Context,
	backendName string,
	model string,
	msgs ...types.Message,
) (chat types.Conversation, err error) {
	backend, defaultModel, err := aigen.loadBackend(ctx, backendName)
	if err != nil {
		return chat, fmt.Errorf("failed loading backend: %w", err)
	}

	if model == "" {
		if defaultModel == "" {
			return nil, types.ErrNoDefaultModel
		}
		model = defaultModel
	}

	return backend.Chat(model, msgs...), nil
}

func (aigen *AiGen) loadBackend(ctx context.Context, name string) (
	backend types.Backend,
	defaultModel string,
	err error,
) {
	if name == "" {
		if aigen.Conf.DefaultBackend == "" {
			return nil, defaultModel, types.ErrNoDefaultBackend
		}
		name = aigen.Conf.DefaultBackend
	}

	// Check if we've already loaded it before
	if backend, ok := aigen.Backends[name]; ok {
		return backend, defaultModel, nil
	}

	// We haven't, check if it's in the configuration
	backendConf, ok := aigen.Conf.Backends[name]
	if !ok {
		return backend, defaultModel, types.ErrNoSuchBackend
	}

	switch backendConf.Type {
	default:
		opts := &openai.Options{
			ApiKey:       derefString(backendConf.APIKey),
			URL:          derefString(backendConf.URL),
			APIVersion:   derefString(backendConf.APIVersion),
			ExtraHeaders: backendConf.ExtraHeaders,
		}
		// default to openai
		backend, err = openai.New(opts)
		if err != nil {
			return nil, defaultModel, err
		}
	}

	return backend, backendConf.DefaultModel, nil
}

// Helper function to get a nilable string
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
