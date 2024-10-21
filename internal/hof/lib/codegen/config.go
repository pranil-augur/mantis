/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/gofireflyio/aiac, licensed under the MIT License.
 */
package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// Config holds the configuration for Mantis.
type Config struct {
	DefaultBackend string                   `json:"default_backend"`
	Backends       map[string]BackendConfig `json:"backends"`
}

// BackendConfig holds backend-specific configuration.
type BackendConfig struct {
	Type         string            `json:"type"`
	APIKey       *string           `json:"api_key,omitempty"`
	DefaultModel string            `json:"default_model"`
	URL          *string           `json:"url,omitempty"`
	APIVersion   *string           `json:"api_version,omitempty"`
	ExtraHeaders map[string]string `json:"extra_headers,omitempty"`
	AWSProfile   *string           `json:"aws_profile,omitempty"`
	AWSRegion    *string           `json:"aws_region,omitempty"`
}

// LoadConfig loads a Mantis configuration file from the provided path, which
// must be a CUE file. If path is an empty string, the default path
// ~/.mantis/config.cue will be used.
func LoadConfig(path string) (Config, error) {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("failed to get user home directory: %w", err)
		}
		path = filepath.Join(homeDir, ".mantis", "config.cue")
	}

	ctx := cuecontext.New()

	instances := load.Instances([]string{path}, nil)
	if len(instances) == 0 {
		return Config{}, fmt.Errorf("no configuration file found at %s", path)
	}

	if instances[0].Err != nil {
		return Config{}, fmt.Errorf("failed to load configuration: %w", instances[0].Err)
	}

	value := ctx.BuildInstance(instances[0])
	if value.Err() != nil {
		return Config{}, fmt.Errorf("failed to build configuration: %w", value.Err())
	}

	var config Config
	if err := value.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("failed to decode configuration: %w", err)
	}

	return config, nil
}
