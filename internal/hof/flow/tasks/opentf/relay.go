/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package opentf

import (
	"fmt"

	"cuelang.org/go/cue"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

// TFTask is a task for running a Terraform plan using a specific configuration
type RelayTask struct {
}

func NewRelayTask(val cue.Value) (hofcontext.Runner, error) {
	return &RelayTask{}, nil
}

func (t *RelayTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value
	script := v.LookupPath(cue.ParsePath("config"))

	// Marshal the unified result to JSON
	jsonScript, err := script.MarshalJSON()
	if err != nil {
		return nil, err
	}
	fmt.Printf("jsonScript %s\n", string(jsonScript))

	exports := v.LookupPath(cue.ParsePath("exports"))

	/*
			   exports: [{
		            var:  "selected_subnet_ids"
		        }]
	*/

	// Extract and parse exports
	var exportedVars []string
	iter, err := exports.List()
	if err != nil {
		return nil, fmt.Errorf("failed to iterate over exports: %v", err)
	}

	for iter.Next() {
		export := iter.Value()
		varName, err := export.LookupPath(cue.ParsePath("var")).String()
		if err != nil {
			return nil, fmt.Errorf("failed to extract var name: %v", err)
		}
		exportedVars = append(exportedVars, varName)
	}

	// Process the script and extract values for exported vars
	for _, varName := range exportedVars {
		value := script.LookupPath(cue.ParsePath(varName))
		if value.Exists() {
			var extracted interface{}
			if err := value.Decode(&extracted); err != nil {
				return nil, fmt.Errorf("failed to decode value for %s: %v", varName, err)
			}
			// add them to the GlobalVars
			ctx.GlobalVars[varName] = extracted
		} else {
			return nil, fmt.Errorf("variable %s not found in script", varName)
		}
	}

	return script, nil
}
