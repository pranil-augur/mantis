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

type Stdout struct{}

func NewStdout(val cue.Value) (hofcontext.Runner, error) {
	return &Stdout{}, nil
}

func (T *Stdout) Run(ctx *hofcontext.Context) (interface{}, error) {
	bufStdout := bufio.NewWriter(ctx.Stdout)
	defer bufStdout.Flush()

	v := ctx.Value
	var m string

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		msg := v.LookupPath(cue.ParsePath("text"))
		if msg.Err() != nil {
			return msg.Err()
		} else if msg.Exists() {
			// Handle string messages
			if str, err := msg.String(); err == nil {
				m = str
			} else {
				// Handle objects/structured data
				syntax := msg.Syntax()
				if syntax != nil {
					m = fmt.Sprintf("%v", syntax)
				} else {
					return fmt.Errorf("failed to convert value to string or object: %v", err)
				}
			}
		} else {
			// Try to print the entire value if 'text' field doesn't exist
			syntax := v.Syntax()
			if syntax != nil {
				m = fmt.Sprintf("%v", syntax)
			} else {
				return fmt.Errorf("no printable value found")
			}
		}
		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	fmt.Fprint(bufStdout, m)
	return nil, nil
}
