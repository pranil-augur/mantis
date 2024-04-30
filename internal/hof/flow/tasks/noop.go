/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package tasks

import (
	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type Noop struct{}

func NewNoop(val cue.Value) (hofcontext.Runner, error) {
	return &Noop{}, nil
}

func (T *Noop) Run(ctx *hofcontext.Context) (interface{}, error) {
	return nil, nil
}
