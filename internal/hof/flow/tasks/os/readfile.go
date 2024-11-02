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
	"fmt"
	g_os "os"

	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type ReadFile struct{}

func NewReadFile(val cue.Value) (hofcontext.Runner, error) {
	return &ReadFile{}, nil
}

func (T *ReadFile) Run(ctx *hofcontext.Context) (interface{}, error) {

	v := ctx.Value
	var fn string
	var err error

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()
		f := v.LookupPath(cue.ParsePath("filename"))

		fn, err = f.String()
		return err
	}()
	if ferr != nil {
		return nil, ferr
	}

	bs, err := g_os.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var res cue.Value
	ferr = func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		c := v.LookupPath(cue.ParsePath("contents"))

		// switch on c's type to fill appropriately
		switch k := c.IncompleteKind(); k {
		case cue.StringKind:
			res = v.FillPath(cue.ParsePath("contents"), string(bs))
		case cue.BytesKind:
			res = v.FillPath(cue.ParsePath("contents"), bs)

		case cue.StructKind:
			ctx := v.Context()
			c := ctx.CompileBytes(bs)
			if c.Err() != nil {
				return c.Err()
			}
			res = v.FillPath(cue.ParsePath("contents"), c)

		case cue.BottomKind:
			res = v.FillPath(cue.ParsePath("contents"), string(bs))

		default:
			return fmt.Errorf("Unsupported Content type in ReadFile task: %q", k)
		}
		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	return res, nil
}
