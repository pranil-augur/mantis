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
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package info

import (
	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type BookkeepingConfig struct {
	Workdir string
}

type Bookkeeping struct {
	cfg  BookkeepingConfig
	val  cue.Value
	next hofcontext.Runner
}

func NewBookkeeping(cfg BookkeepingConfig, opts flags.RootPflagpole, popts flags.FlowPflagpole) *Bookkeeping {
	return &Bookkeeping{
		cfg: cfg,
	}
}

func (M *Bookkeeping) Run(ctx *hofcontext.Context) (results interface{}, err error) {
	// bt := ctx.BaseTask
	// fmt.Println("bt:", bt.ID, bt.UUID)
	result, err := M.next.Run(ctx)

	// write out file in background
	return result, err
}

func (M *Bookkeeping) Apply(ctx *hofcontext.Context, runner hofcontext.RunnerFunc) hofcontext.RunnerFunc {
	return func(val cue.Value) (hofcontext.Runner, error) {
		// id := fmt.Sprint(val.Path())
		// fmt.Println("book: found @", val.Path())
		next, err := runner(val)
		if err != nil {
			return nil, err
		}
		return &Bookkeeping{
			val:  val,
			next: next,
		}, nil
	}
}

func (M *Bookkeeping) write(filename string, val cue.Value) error {

	return nil
}
