/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package helpers

import (
	"fmt"
	"time"
)

func (sc *SourceConfig) Encode() (map[string]any, error) {
	return map[string]any{
		"name":   sc.Name,
		"source": sc.Source,
		"args": sc.Args,
		"watch": sc.WatchTime.String(),
		"globs": sc.WatchGlobs,
	}, nil
}

func (sc *SourceConfig) Decode(input map[string]any) (*SourceConfig, error) {
	aargs := input["args"].([]any)
	args := make([]string, len(aargs))
	for i, a := range aargs {
		args[i] = a.(string)
	}
	watch := "0s"
	if _w, ok := input["watch"]; ok {
		watch = _w.(string)
	}
	d, err := time.ParseDuration(watch)
	if err != nil {
		return nil, fmt.Errorf("error while decoding SourceConfig.watch: %w", err)
	}

	var globs []string
	if _gs, ok := input["globs"]; ok {
		gs := _gs.([]any)
		for _, g := range gs {
			globs = append(globs, g.(string))
		}
	}

	var n string
	if _n, ok := input["name"]; ok {
		n = _n.(string)
	}

	return &SourceConfig{
		Name: n,
		Source: EvalSource(input["source"].(string)),
		Args: args,
		WatchTime: d,
		WatchGlobs: globs,
	}, nil
}

