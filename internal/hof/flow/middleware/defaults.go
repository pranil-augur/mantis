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

package middleware

import (
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"

	"github.com/opentofu/opentofu/internal/hof/flow/middleware/info"
	"github.com/opentofu/opentofu/internal/hof/flow/middleware/sync"
)

func UseDefaults(ctx *hofcontext.Context, opts flags.RootPflagpole, popts flags.FlowPflagpole) {
	// ctx.Use(info.NewPrint(opts, popts))
	ctx.Use(info.NewProgress(opts, popts))
	//ctx.Use(info.NewBookkeeping(info.BookkeepingConfig{
	//Workdir: ".hof/flow",
	//}, opts, popts))
	ctx.Use(sync.NewPool(opts, popts))
}
