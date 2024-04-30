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
	"bufio"
	"fmt"

	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type Stdin struct{}

func NewStdin(val cue.Value) (hofcontext.Runner, error) {
	return &Stdin{}, nil
}

func (T *Stdin) Run(ctx *hofcontext.Context) (interface{}, error) {

	v := ctx.Value
	var m string

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()
		var err error

		msg := v.LookupPath(cue.ParsePath("msg"))
		if msg.Err() != nil {
			return msg.Err()

		} else if msg.Exists() {
			m, err = msg.String()
			if err != nil {
				return err
			}
			// print message to user
		}
		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	if len(m) > 0 {
		fmt.Fprint(ctx.Stdout, m)
	}
	reader := bufio.NewReader(ctx.Stdin)
	text, _ := reader.ReadString('\n')

	ctx.CUELock.Lock()
	defer ctx.CUELock.Unlock()
	res := v.FillPath(cue.ParsePath("contents"), text)

	return res, nil
}
