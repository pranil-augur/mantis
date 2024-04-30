/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package os

import (
	"os"

	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type Mkdir struct{}

func NewMkdir(val cue.Value) (hofcontext.Runner, error) {
	return &Mkdir{}, nil
}

func (T *Mkdir) Run(ctx *hofcontext.Context) (interface{}, error) {

	v := ctx.Value

	var dir string
	var err error

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		d := v.LookupPath(cue.ParsePath("dir"))
		if d.Err() != nil {
			return d.Err()
		} else if d.Exists() {
			dir, err = d.String()
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	// TODO, make mode configurable
	err = os.MkdirAll(dir, 0755)

	return nil, err
}
