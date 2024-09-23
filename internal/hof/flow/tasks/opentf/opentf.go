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
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/command"
	"github.com/opentofu/opentofu/internal/command/cliconfig"
	"github.com/opentofu/opentofu/internal/configs"
	"github.com/opentofu/opentofu/internal/getproviders"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/terminal"
	"github.com/opentofu/opentofu/internal/utils"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// TFTask is a task for running a Terraform plan using a specific configuration
type TFTask struct {
}

// Assuming Ui is a global variable of type cli.Ui
var Ui cli.Ui

func NewTFTask(val cue.Value) (hofcontext.Runner, error) {
	return &TFTask{}, nil
}

func (t *TFTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value
	script := v.LookupPath(cue.ParsePath("config"))

	// Marshal the unified result to JSON
	jsonScript, err := script.MarshalJSON()

	// Print the JSON representation of the script
	if err != nil {
		return nil, fmt.Errorf("error marshalling script to JSON: %v", err)
	}
	scriptStr := string(jsonScript)
	if err != nil {
		return nil, fmt.Errorf("error retrieving script as string: %v", err)
	}
	// Serialize JSON string to bytes
	scriptBytes := []byte(scriptStr)
	if len(scriptBytes) == 0 {
		return nil, fmt.Errorf("serialized JSON is empty")
	}
	// Load configuration
	config, diags := cliconfig.LoadConfig()
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load CLI configuration: %v", diags.Err())
	}

	// Initialize services
	services := disco.NewWithCredentialsSource(nil) // Simplified for example

	// Initialize provider source and overrides
	providerSrc := getproviders.NewRegistrySource(services)
	providerDevOverrides := map[addrs.Provider]getproviders.PackageLocalDir{}

	// Initialize unmanaged providers (simplified)
	unmanagedProviders := map[addrs.Provider]*plugin.ReattachConfig{}

	// Initialize terminal streams
	streams, err := terminal.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize terminal: %v", err)
	}

	var std_ctx context.Context

	taskPath := ctx.BaseTask.ID
	configDetails := &configs.MicroConfig{
		Identifier: taskPath,
		Content:    scriptBytes,
		Format:     "json",
	}
	// Initialize commands
	commandsFactory := utils.InitCommandsWrapper(std_ctx, "", streams, config, services, providerSrc, providerDevOverrides, unmanagedProviders, configDetails)
	if ctx.Plan {
		// Retrieve the 'plan' command from the commandsFactory using the appropriate key
		planCommandFactory, exists := commandsFactory["plan"]
		if !exists {
			return nil, fmt.Errorf("plan command not found in commands factory")
		}

		// Generate the plan command using the factory
		planCommandInterface, err := planCommandFactory()
		if err != nil {
			return nil, fmt.Errorf("error generating plan command: %v", err)
		}

		// Assert the type of the command to *command.PlanCommand
		planCommand, ok := planCommandInterface.(*command.PlanCommand)
		if !ok {
			return nil, fmt.Errorf("error asserting command type to *command.PlanCommand")
		}
		// Execute the PlanCommand with the configuration file path
		// planCommand.Meta.ConfigByteArray = scriptBytes
		var parsedVariables *map[string]map[string]cty.Value = &map[string]map[string]cty.Value{}
		// Create a new TFContext with the parsedVariables
		tfContext := hofcontext.NewTFContext(parsedVariables)

		_, err = planCommand.RunAPI([]string{}, tfContext)
		if err != nil {
			return nil, fmt.Errorf("failed to execute plan command with exit status %d", err)
		}
		parsedVariablesMap, _ := convertCtyValueToMap(*parsedVariables)
		fmt.Printf("Parsed Variables: %+v\n", parsedVariablesMap)
		// v.FillPath(cue.ParsePath("out"), parsedVariables)
		// Attempt to fill the path with the new value
		newV := v.FillPath(cue.ParsePath("out"), parsedVariablesMap)
		// verifyNewV(newV)

		return newV, nil
	} else if ctx.Apply || ctx.Destroy {
		var applyCommandFactory cli.CommandFactory
		var exists bool

		if ctx.Apply {
			// Retrieve the 'plan' command from the commandsFactory using the appropriate key
			applyCommandFactory, exists = commandsFactory["apply"]
		} else {
			applyCommandFactory, exists = commandsFactory["destroy"]
		}

		if !exists {
			return nil, fmt.Errorf("apply command not found in commands factory")
		}

		// Generate the plan command using the factory
		applyCommandInterface, err := applyCommandFactory()
		if err != nil {
			return nil, fmt.Errorf("error generating apply command: %v", err)
		}

		// Assert the type of the command to *command.PlanCommand
		applyCommand, ok := applyCommandInterface.(*command.ApplyCommand)
		if !ok {
			return nil, fmt.Errorf("error asserting command type to *command.ApplyCommand")
		}
		// Execute the PlanCommand with the configuration file path
		// planCommand.Meta.ConfigByteArray = scriptBytes
		var parsedVariables *map[string]map[string]cty.Value = &map[string]map[string]cty.Value{}
		// Create a new TFContext with the parsedVariables
		tfContext := hofcontext.NewTFContext(parsedVariables)

		_, err = applyCommand.RunAPI([]string{}, tfContext)
		if err != nil {
			return nil, fmt.Errorf("failed to execute apply command with exit status %d", err)
		}
		v.FillPath(cue.ParsePath("out"), parsedVariables)
		// Attempt to fill the path with the new value
		newV := v.FillPath(cue.ParsePath("out"), parsedVariables)

		return newV, nil
	} else if ctx.Init {
		cueContext := cuecontext.New()
		value := cueContext.CompileString(scriptStr, cue.Filename("input.json"))
		terraformOrModule := value.LookupPath(cue.ParsePath("terraform")).Exists() || value.LookupPath(cue.ParsePath("module")).Exists()
		if !terraformOrModule {
			return nil, nil
		}

		initCommandFactory, exists := commandsFactory["init"]

		if !exists {
			return nil, fmt.Errorf("init command not found in commands factory")
		}

		// Generate the plan command using the factory
		initCommandInterface, err := initCommandFactory()
		if err != nil {
			return nil, fmt.Errorf("error generating init command: %v", err)
		}

		// Assert the type of the command to *command.PlanCommand
		initCommand, ok := initCommandInterface.(*command.InitCommand)
		if !ok {
			return nil, fmt.Errorf("error asserting command type to *command.PlanCommand")
		}

		retval := initCommand.Run([]string{})
		if retval < 0 {
			return nil, fmt.Errorf("error Initializing")
		}
	} else {
		return nil, fmt.Errorf("unknown command. Need to use one of plan/apply/init/destroy")
	}
	return nil, nil
}

func convertCtyValueToMap(input map[string]map[string]cty.Value) (map[string]interface{}, error) {
	// Convert the entire input to a cty.Value
	inputValue, err := convertMapToCtyValue(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input to cty.Value: %v", err)
	}

	// Marshal cty.Value to JSON, preserving type information
	jsonBytes, err := ctyjson.Marshal(inputValue, inputValue.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cty.Value to JSON: %v", err)
	}

	// Unmarshal JSON back to cty.Value, using the original type
	ctyValue, err := ctyjson.Unmarshal(jsonBytes, inputValue.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to cty.Value: %v", err)
	}

	// Convert cty.Value to Go map
	return ctyValueToMap(ctyValue)
}

func convertMapToCtyValue(input map[string]map[string]cty.Value) (cty.Value, error) {
	outerMap := make(map[string]cty.Value)
	for outerKey, innerMap := range input {
		innerCtyMap := make(map[string]cty.Value)
		for innerKey, value := range innerMap {
			innerCtyMap[innerKey] = value
		}
		outerMap[outerKey] = cty.ObjectVal(innerCtyMap)
	}
	return cty.ObjectVal(outerMap), nil
}

func ctyValueToMap(v cty.Value) (map[string]interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	if !v.Type().IsObjectType() && !v.Type().IsMapType() {
		return nil, fmt.Errorf("cannot convert non-object/non-map value to map")
	}
	result := make(map[string]interface{})
	for key, value := range v.AsValueMap() {
		var err error
		result[key], err = ctyValueToInterface(value)
		if err != nil {
			return nil, fmt.Errorf("error converting value for key %s: %v", key, err)
		}
	}
	return result, nil
}

func ctyValueToInterface(v cty.Value) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	switch {
	case v.Type() == cty.String:
		return v.AsString(), nil
	case v.Type() == cty.Number:
		return v.AsBigFloat(), nil
	case v.Type() == cty.Bool:
		return v.True(), nil
	case v.Type().IsListType() || v.Type().IsSetType() || v.Type().IsTupleType():
		list := make([]interface{}, 0, v.LengthInt())
		for _, ev := range v.AsValueSlice() {
			elemValue, err := ctyValueToInterface(ev)
			if err != nil {
				return nil, err
			}
			list = append(list, elemValue)
		}
		return list, nil
	case v.Type().IsMapType() || v.Type().IsObjectType():
		return ctyValueToMap(v)
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// Debug function to print cty.Value
func printCtyValue(v cty.Value) {
	fmt.Printf("Type: %s\n", v.Type().FriendlyName())
	fmt.Printf("Value: %#v\n", v)
}

// func convertCtyValueToMap(input map[string]map[string]cty.Value) map[string]interface{} {
// 	result := make(map[string]interface{})

// 	for outerKey, innerMap := range input {
// 		innerResult := make(map[string]interface{})
// 		for innerKey, ctyValue := range innerMap {
// 			innerResult[innerKey] = convertCtyValueToInterface(ctyValue)
// 		}
// 		result[outerKey] = innerResult
// 	}

// 	return result
// }

// func convertCtyValueToInterface(v cty.Value) interface{} {
// 	if v.IsNull() {
// 		return nil
// 	}

// 	switch v.Type() {
// 	case cty.String:
// 		return v.AsString()
// 	case cty.Number:
// 		f, _ := v.AsBigFloat().Float64()
// 		return f
// 	case cty.Bool:
// 		return v.True()
// 	case cty.List(cty.DynamicPseudoType), cty.Set(cty.DynamicPseudoType):
// 		list := make([]interface{}, 0, v.LengthInt())
// 		for _, ev := range v.AsValueSlice() {
// 			list = append(list, convertCtyValueToInterface(ev))
// 		}
// 		return list
// 	case cty.Map(cty.DynamicPseudoType), cty.Object(map[string]cty.Type{}):
// 		m := make(map[string]interface{})
// 		for k, ev := range v.AsValueMap() {
// 			m[k] = convertCtyValueToInterface(ev)
// 		}
// 		return m
// 	default:
// 		// For other types, convert to string representation
// 		return fmt.Sprintf("%v", v)
// 	}
// }
