/* Copyright 2024 Augur AI
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 */

package mantis

import (
	"fmt"
	"log"

	"cuelang.org/go/cue"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

func evaluate(ctx *cue.Context, node *cue.Value) (*cue.Value, error) {
	return node, nil
}

type LocalEvaluator struct{}

func NewLocalEvaluator(val cue.Value) (hofcontext.Runner, error) {
	return &LocalEvaluator{}, nil
}

// Run processes locals and dynamically evaluates expressions
func (T *LocalEvaluator) Run(ctx *hofcontext.Context) (interface{}, error) {

	v := ctx.Value

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		// Input will be in the form of :
		// locals : {
		// 	@task(mantis.evaluate)
		// 		exports: [{
		// 			expression: string
		// 			alias: string
		// 		}]
		// }

		exports := v.LookupPath(cue.ParsePath("exports"))
		iter, _ := exports.List()
		for iter.Next() {
			cueExpression := iter.Value().LookupPath(cue.ParsePath("cueexpr"))
			if !cueExpression.Exists() {
				log.Fatalf("Path 'cueexpr' not found in CUE file")
			}
			varVal := iter.Value().LookupPath(cue.ParsePath("var"))
			varStr, err := varVal.String()
			if err != nil {
				return err
			}
			exprStr, err := cueExpression.String()
			if err != nil {
				return err
			}
			evalContext := v.Context()
			transformedValue := evalContext.CompileString(exprStr, cue.Scope(ctx.RootValue))
			if transformedValue.Err() != nil {
				fmt.Printf("Failed to compile CUE expression: %v\n", transformedValue.Err())
			}
			fmt.Printf("Transformed value: %v\n", transformedValue)
			// Set the transformed value in the global vars
			ctx.GlobalVars[varStr] = transformedValue
		}
		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	// Return updated CUE context with evaluated locals
	return ctx.Value, nil
}
