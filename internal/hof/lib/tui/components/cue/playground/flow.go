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

package playground

import (
	"bytes"

	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	flowcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/flow/middleware"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks"
	"github.com/opentofu/opentofu/internal/hof/flow/flow"
)


func (C *Playground) runFlow(val cue.Value) (cue.Value, error) {
	var stdin, stdout, stderr bytes.Buffer

	ctx := flowcontext.New()
	ctx.RootValue = val
	ctx.Stdin = &stdin
	ctx.Stdout = &stdout
	ctx.Stderr = &stderr

	// how to inject tags into original value
	// fill / return value
	middleware.UseDefaults(ctx, flags.RootPflagpole{}, flags.FlowPflagpole{})
	tasks.RegisterDefaults(ctx)

	f, err := flow.OldFlow(ctx, val)
	if err != nil {
		return val, err
	}

	C.flow = f

	err = f.Start()

	return f.Final, err
}
