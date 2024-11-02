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

package sync

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type Wait struct {
	// "required"
	val  cue.Value
	next hofcontext.Runner

	// local
	name string
}

func NewWait(opts flags.RootPflagpole, popts flags.FlowPflagpole) *Wait {
	return &Wait{}
}

func (M *Wait) Run(ctx *hofcontext.Context) (results interface{}, err error) {
	// wg := ctx.<middleware>.[M.name]
	// wg.Add(1)
	// defer wg.Done()

	fmt.Println("wait: pre @", M.val.Path())
	// should this happen during discovery? (in Apply)
	result, err := M.next.Run(ctx)
	fmt.Println("wait: post @", M.val.Path())

	return result, err
}

func (M *Wait) Apply(ctx *hofcontext.Context, runner hofcontext.RunnerFunc) hofcontext.RunnerFunc {
	return func(val cue.Value) (hofcontext.Runner, error) {
		hasAttr := false
		attrs := val.Attributes(cue.ValueAttr)
		var a cue.Attribute
		for _, attr := range attrs {
			if attr.Name() == "wait" {
				a = attr
				hasAttr = true
				break
			}
		}

		next, err := runner(val)
		if err != nil {
			return nil, err
		}

		if !hasAttr {
			return next, nil
		}

		fmt.Println("wait: found @", val.Path(), a)

		// what is in this attribute

		// setup wait by name here

		return &Wait{
			// required
			val:  val,
			next: next,
			// extra
			// name:
		}, nil
	}
}
