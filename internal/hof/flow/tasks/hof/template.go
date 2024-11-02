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

package hof

import (
	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"

	"github.com/opentofu/opentofu/internal/hof/lib/templates"
)

type HofTemplate struct {
	Name     string
	Data     any
	Template string
	Partials map[string]string

	Delims templates.Delims
}

func NewHofTemplate(val cue.Value) (hofcontext.Runner, error) {
	return &HofTemplate{}, nil
}

func (T *HofTemplate) Run(ctx *hofcontext.Context) (interface{}, error) {

	v := ctx.Value

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		err := v.Decode(T)

		return err
	}()
	if ferr != nil {
		return nil, ferr
	}

	t, err := templates.CreateFromString(T.Name, T.Template, T.Delims)
	if err != nil {
		return nil, err
	}

	for k, P := range T.Partials {
		p := t.T.New(k)
		// do we need to do this, does the partial use the helpers already registered?
		// T.AddGolangHelpers()
		_, err := p.Parse(P)
		if err != nil {
			return nil, err
		}
	}

	bs, err := t.Render(T.Data)
	if err != nil {
		return nil, err
	}

	ctx.CUELock.Lock()
	defer ctx.CUELock.Unlock()
	res := v.FillPath(cue.ParsePath("out"), string(bs))

	return res, nil
}
