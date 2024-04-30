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
	"fmt"

	"cuelang.org/go/cue"

	flowctx "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/flow/flow"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

// this is buggy, need upstream support
type Nest struct{}

func NewNest(val cue.Value) (flowctx.Runner, error) {
	return &Nest{}, nil
}

func (T *Nest) Run(ctx *flowctx.Context) (interface{}, error) {
	val := ctx.Value

	orig := ctx.FlowStack
	ctx.FlowStack = append(orig, fmt.Sprint(val.Path()))

	n, err := hof.ParseHof[flow.Flow](val)
	if err != nil {
		return nil, err
	}

	p, err := flow.OldFlow(ctx, val)
	if err != nil {
		return nil, err
	}

	p.Node = n

	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("in nested task: %w", err)
	}

	ctx.FlowStack = orig

	return p.Final, nil
}
