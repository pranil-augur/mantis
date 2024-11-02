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
	g_os "os"

	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
)

type ReadGlobs struct{}

func NewReadGlobs(val cue.Value) (hofcontext.Runner, error) {
	return &ReadGlobs{}, nil
}

func (T *ReadGlobs) Run(ctx *hofcontext.Context) (interface{}, error) {

	val := ctx.Value

	patterns, err := extractGlobConfig(ctx, val)
	if err != nil {
		return nil, err
	}

	filepaths, err := yagu.FilesFromGlobs(patterns)
	if err != nil {
		return nil, err
	}

	data := map[string]string{}

	for _, file := range filepaths {
		bs, err := g_os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		data[file] = string(bs)
	}

	return map[string]interface{}{"files": data}, nil
}
