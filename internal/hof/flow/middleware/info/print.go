/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package info

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type Print struct {
	val  cue.Value
	next hofcontext.Runner
}

func NewPrint(opts flags.RootPflagpole, popts flags.FlowPflagpole) *Print {
	return &Print{}
}

func (M *Print) Run(ctx *hofcontext.Context) (results interface{}, err error) {
	result, err := M.next.Run(ctx)

	fmt.Println("HELLO")

	attrs := M.val.Attributes(cue.ValueAttr)
	for _, attr := range attrs {
		if attr.Name() == "print" {
			for i := 0; i < attr.NumArgs(); i++ {
				p, err := attr.String(i)
				if err != nil {
					return result, err
				}
				if p == "" {
					fmt.Printf("%s: %v\n", M.val.Path(), result)
				} else {
					r, ok := result.(cue.Value)
					var v cue.Value
					if ok {
						v = r.LookupPath(cue.ParsePath(p))
					} else {
						v = M.val.LookupPath(cue.ParsePath(p))
					}
					fmt.Printf("%s: %v\n", v.Path(), v)
				}
			}
			break
		}
	}

	return result, err
}

func (M *Print) Apply(ctx *hofcontext.Context, runner hofcontext.RunnerFunc) hofcontext.RunnerFunc {
	return func(val cue.Value) (hofcontext.Runner, error) {
		// parse out the local #hof node data
		node, err := hof.ParseHof[any](val)
		if err != nil {
			return nil, err
		}
		if node == nil {
			panic("we should have found a node to even get here")
		}

		p := node.Hof.Flow.Print

		// convenience
		fmt.Println("found print", val.Path(), node.Hof.Path, p)

		next, err := runner(val)
		if err != nil {
			return nil, err
		}

		if p.Level > 0 {
			return next, nil
		}

		fmt.Println("found print")

		return &Print{
			val:  val,
			next: next,
		}, nil
	}
}
