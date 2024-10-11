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
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type LocalEvaluator struct{}

func NewLocalEvaluator(val cue.Value) (hofcontext.Runner, error) {
	return &LocalEvaluator{}, nil
}

// Run processes locals and dynamically evaluates expressions
func (T *LocalEvaluator) Run(ctx *hofcontext.Context) (interface{}, error) {
	v := ctx.Value

	ferr := func() error {
		ctx.CUELock.Lock()
		defer ctx.CUELock.Unlock()

		exports := v.LookupPath(cue.ParsePath("exports"))
		iter, _ := exports.List()

		for iter.Next() {
			cueExpression := iter.Value().LookupPath(cue.ParsePath("cueexpr"))
			if !cueExpression.Exists() {
				return fmt.Errorf("path 'cueexpr' not found in CUE file")
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

			// Create a temporary CUE file with the expression and necessary imports
			tmpFile, err := createTempCueFile(exprStr)
			if err != nil {
				return fmt.Errorf("failed to create temporary CUE file: %w", err)
			}
			defer os.Remove(tmpFile)

			// Create a new CUE context
			cueCtx := cuecontext.New()

			// Load the CUE file
			instances := load.Instances([]string{tmpFile}, nil)
			if len(instances) == 0 {
				return fmt.Errorf("no instances loaded")
			}

			// Build the CUE value
			value := cueCtx.BuildInstance(instances[0])
			if value.Err() != nil {
				return fmt.Errorf("failed to build CUE instance: %w", value.Err())
			}

			// Evaluate the expression in the context of the root value
			transformedValue := value.FillPath(cue.Path{}, ctx.RootValue)
			if transformedValue.Err() != nil {
				return fmt.Errorf("failed to evaluate CUE expression: %w", transformedValue.Err())
			}

			// Set the transformed value in the global vars
			ctx.GlobalVars.Store(varStr, transformedValue)
		}
		return nil
	}()

	if ferr != nil {
		return nil, ferr
	}

	// Return updated CUE context with evaluated locals
	return ctx.Value, nil
}

func createTempCueFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "evaluate*.cue")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
