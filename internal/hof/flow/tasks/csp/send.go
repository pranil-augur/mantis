/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package csp

import (
	"fmt"

	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type Send struct{}

func NewSend(val cue.Value) (hofcontext.Runner, error) {
	return &Send{}, nil
}

func (T *Send) Run(ctx *hofcontext.Context) (interface{}, error) {
	// fmt.Println("csp.Send", ctx.Value)

	v := ctx.Value
	var (
		err     error
		mailbox string
		key     string
		val     cue.Value
	)

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		val = v.LookupPath(cue.ParsePath("val"))
		if !val.Exists() {
			return fmt.Errorf("in csp.Send task %s: missing field 'val'", v.Path())
		}
		if val.Err() != nil {
			return val.Err()
		}
		// fmt.Println("csp.Send().val:", val)

		kv := v.LookupPath(cue.ParsePath("key"))
		if kv.Exists() {
			if kv.Err() != nil {
				return kv.Err()
			}
			key, err = kv.String()
			if err != nil {
				return err
			}
		}

		nv := v.LookupPath(cue.ParsePath("mailbox"))
		if !nv.Exists() {
			return fmt.Errorf("in csp.Send task %s: missing field 'mailbox'", v.Path())
		}
		if nv.Err() != nil {
			return nv.Err()
		}
		mailbox, err = nv.String()
		if err != nil {
			return err
		}

		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	// load mailbox
	// fmt.Println("mailbox?:", mailbox)
	ci, loaded := ctx.Mailbox.Load(mailbox)
	if !loaded {
		return nil, fmt.Errorf("channel %q not found", mailbox)
	}

	msg := Msg{
		Key: key,
		Val: val,
	}
	// fmt.Println("sending:", msg)
	// send a Msg
	c := ci.(chan Msg)
	c <- msg

	return nil, nil
}
