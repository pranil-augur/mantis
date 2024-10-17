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
	"regexp"
	"strings"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type LocalEvaluator struct{}

func NewLocalEvaluator(val cue.Value) (hofcontext.Runner, error) {
	return &LocalEvaluator{}, nil
}

func FormatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v) // Quote string values
	case []interface{}:
		// Handle arrays of interface{}
		formattedElements := make([]string, len(v))
		for i, elem := range v {
			formattedElements[i] = FormatValue(elem) // Recursively format each element
		}
		return fmt.Sprintf("[%s]", strings.Join(formattedElements, ", "))
	case int, float64, bool:
		// Return numbers and booleans without quotes
		return fmt.Sprintf("%v", v)
	default:
		// Handle any other types with default formatting
		return fmt.Sprintf("%v", v)
	}
}
func PopulateTemplate(template string, vars *sync.Map) (string, error) {
	// Define a regex pattern to match @var(id_here)
	re := regexp.MustCompile(`@var\((\w+)\)`)

	// Function to replace each match with the corresponding value from the sync.Map
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract the variable name using capturing groups
		matches := re.FindStringSubmatch(match)
		if len(matches) > 1 {
			varName := matches[1] // This will be the captured variable name

			// Load the value from sync.Map
			if value, exists := vars.Load(varName); exists {
				// Use the FormatValue function to handle the formatting
				return FormatValue(value)
			}
		}
		return match // If the variable doesn't exist, return the original match
	})
	return result, nil
}

// Run processes locals and dynamically evaluates expressions
func (T *LocalEvaluator) Run(ctx *hofcontext.Context) (interface{}, error) {
	v := ctx.Value
	if !ctx.Apply {
		return v, nil
	}
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

			exprStr, err := cueExpression.String()
			if err != nil {
				return err
			}

			populatedCueExpr, err := PopulateTemplate(exprStr, ctx.GlobalVars)
			if err != nil {
				return fmt.Errorf("failed to update variables in CUE file: %w", err)
			}
			// Create a temporary CUE file with the expression and necessary imports
			tmpFile, err := createTempCueFile(populatedCueExpr)
			if err != nil {
				return fmt.Errorf("failed to create temporary CUE file: %w", err)
			}
			defer os.Remove(tmpFile)

			// Load the CUE file
			instances := load.Instances([]string{tmpFile}, nil)
			if len(instances) == 0 {
				return fmt.Errorf("no instances loaded")
			}

			// Build the CUE value
			value := ctx.CueContext.BuildInstance(instances[0])
			if value.Err() != nil {
				return fmt.Errorf("failed to build CUE instance: %w", value.Err())
			}

			result := value.LookupPath(cue.ParsePath("result"))
			// fmt.Println("Result is: ")
			// fmt.Printf(result.String())
			v = v.FillPath(cue.ParsePath("out"), result)
			ctx.Value = v
			// fmt.Println("ctx updated is: ")
			// fmt.Printf(ctx.Value.String())

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
