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

package prompt

import (
	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	libprompt "github.com/opentofu/opentofu/internal/hof/lib/prompt"
)

type Prompt struct{}

func NewPrompt(val cue.Value) (hofcontext.Runner, error) {
	return &Prompt{}, nil
}

// Tasks must implement a Run func, this is where we execute our task
func (T *Prompt) Run(ctx *hofcontext.Context) (any, error) {
	ctx.CUELock.Lock()
	defer ctx.CUELock.Unlock()

	r, err := libprompt.RunPrompt(ctx.Value)
	if err != nil {
		return nil, err
	}

	return r, nil
}
