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

package os

import (
	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
)

type Glob struct{}

func NewGlob(val cue.Value) (hofcontext.Runner, error) {
	return &Glob{}, nil
}

func (T *Glob) Run(ctx *hofcontext.Context) (interface{}, error) {

	val := ctx.Value

	patterns, err := extractGlobConfig(ctx, val)
	if err != nil {
		return nil, err
	}

	filepaths, err := yagu.FilesFromGlobs(patterns)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"filepaths": filepaths}, nil
}

func extractGlobConfig(ctx *hofcontext.Context, val cue.Value) (patterns []string, err error) {
	ctx.CUELock.Lock()
	defer ctx.CUELock.Unlock()

	ps := val.LookupPath(cue.ParsePath("globs"))
	if ps.Err() != nil {
		return nil, ps.Err()
	}

	iter, err := ps.List()
	if err != nil {
		return nil, err
	}

	for iter.Next() {
		gv := iter.Value()
		if gv.Err() != nil {
			return nil, gv.Err()
		}
		gs, err := gv.String()
		if err != nil {
			return nil, err
		}

		patterns = append(patterns, gs)
	}

	return patterns, nil
}
