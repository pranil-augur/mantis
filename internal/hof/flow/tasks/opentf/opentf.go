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
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"github.com/opentofu/opentofu/internal/addrs"
	backendInit "github.com/opentofu/opentofu/internal/backend/init"
	"github.com/opentofu/opentofu/internal/command"
	"github.com/opentofu/opentofu/internal/command/arguments"
	"github.com/opentofu/opentofu/internal/command/cliconfig"
	"github.com/opentofu/opentofu/internal/configs"
	"github.com/opentofu/opentofu/internal/getproviders"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
	"github.com/opentofu/opentofu/internal/terminal"
	"github.com/opentofu/opentofu/internal/utils"
	"github.com/zclconf/go-cty/cty"
)

// TFTask is a task for running a Terraform plan using a specific configuration
type TFTask struct {
}

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
		return nil, fmt.Errorf("error marshalling script to json: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("error appending backend config: %v", err)
	}

	scriptStr := string(jsonScript)

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

	// Initialize the backends.
	backendInit.Init(services)

	var std_ctx context.Context
	backendStatePath :=
		fmt.Sprintf("./terraform/back_%s.tfstate", ctx.BaseTask.ID)

	taskPath := ctx.BaseTask.ID
	configDetails := &configs.MicroConfig{
		Identifier:       taskPath,
		Content:          scriptBytes,
		Format:           "json",
		BackendStatePath: backendStatePath,
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

		// Create and populate the Apply arguments
		planArgs := &arguments.Plan{
			State: &arguments.State{
				StatePath: createStatePath(ctx.BaseTask.ID),
			},
		}

		rawArgs := []string{}

		// Add state path if specified
		if planArgs.State.StatePath != "" {
			rawArgs = append(rawArgs, "-state="+planArgs.State.StatePath)
		}

		// Execute the PlanCommand with the configuration file path
		// planCommand.Meta.ConfigByteArray = scriptBytes
		parsedVariables := sync.Map{}
		// Create a new TFContext with the parsedVariables
		tfContext := hofcontext.NewTFContext(&parsedVariables)

		_, err = planCommand.RunAPI(rawArgs, tfContext)
		if err != nil {
			return nil, fmt.Errorf("failed to execute apply command with exit status %d", err)
		}
		var parsedVariablesMap map[string]interface{}
		parsedVariablesMap, _ = convertCtyToGo(&parsedVariables)
		// fmt.Printf("Parsed Variables: %+v\n", parsedVariablesMap)
		// v.FillPath(cue.ParsePath("out"), parsedVariables)
		// Attempt to fill the path with the new value
		newV := v.FillPath(cue.ParsePath("out"), parsedVariablesMap)

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

		// Create and populate the Apply arguments
		applyArgs := &arguments.Apply{
			State: &arguments.State{
				StatePath: createStatePath(ctx.BaseTask.ID),
			},
		}
		// Execute the PlanCommand with the configuration file path
		// planCommand.Meta.ConfigByteArray = scriptBytes
		// var parsedVariables *map[string]map[string]cty.Value = &map[string]map[string]cty.Value{}
		parsedVariables := sync.Map{}
		// Create a new TFContext with the parsedVariables
		tfContext := hofcontext.NewTFContext(&parsedVariables)

		// Add apply args to rawArgs
		rawArgs := []string{"-auto-approve"}

		// Add state path if specified
		if applyArgs.State.StatePath != "" {
			rawArgs = append(rawArgs, "-state="+applyArgs.State.StatePath)
		}

		_, err = applyCommand.RunAPI(rawArgs, tfContext)
		if err != nil {
			return nil, fmt.Errorf("failed to execute apply command with exit status %d", err)
		}
		var parsedVariablesMap map[string]interface{}
		parsedVariablesMap, _ = convertCtyToGo(&parsedVariables)
		// fmt.Printf("Parsed Variables: %+v\n", parsedVariablesMap)
		// v.FillPath(cue.ParsePath("out"), parsedVariables)
		// Attempt to fill the path with the new value
		newV := v.FillPath(cue.ParsePath(hof.MantisTaskOuts), parsedVariablesMap)

		return newV, nil
	} else if ctx.Init {
		cueContext := cuecontext.New()
		value := cueContext.CompileString(scriptStr, cue.Filename(hof.MantisJsonConfig))
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

		retval := initCommand.Run([]string{"-reconfigure"})
		if retval < 0 {
			return nil, fmt.Errorf("error Initializing")
		}
	} else {
		return nil, fmt.Errorf("unknown command. Need to use one of plan/apply/destroy")
	}
	return nil, nil
}

func createStatePath(taskID string) string {
	return fmt.Sprintf(hof.MantisStateFilePath, taskID)
}

func convertCtyToGo(input *sync.Map) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var conversionError error

	input.Range(func(key, value interface{}) bool {
		convertedValue, err := convertValue(value)
		if err != nil {
			conversionError = fmt.Errorf("error converting key '%v': %w", key, err)
			return false // Stop iteration on error
		}

		// Assuming the key is a string, if not, you may need to convert it
		strKey, ok := key.(string)
		if !ok {
			conversionError = fmt.Errorf("key is not a string: %v", key)
			return false // Stop iteration
		}

		result[strKey] = convertedValue
		return true // Continue iteration
	})

	if conversionError != nil {
		return nil, conversionError
	}

	return result, nil
}

func convertValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case cty.Value:
		return ctyValueToGo(v)
	case map[string]cty.Value:
		return convertCtyValueMap(v)
	case sync.Map:
		return convertCtyToGo(&v)
	case *sync.Map:
		return convertCtyToGo(v)
	case []interface{}:
		return convertSlice(v)
	default:
		// If it's not a recognized type, return it as-is
		return v, nil
	}
}

func convertCtyValueMap(input map[string]cty.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range input {
		convertedValue, err := ctyValueToGo(value)
		if err != nil {
			return nil, fmt.Errorf("error converting key '%s': %w", key, err)
		}
		result[key] = convertedValue
	}
	return result, nil
}

func convertSlice(slice []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		convertedValue, err := convertValue(v)
		if err != nil {
			return nil, fmt.Errorf("error converting slice element at index %d: %w", i, err)
		}
		result[i] = convertedValue
	}
	return result, nil
}

func ctyValueToGo(v cty.Value) (interface{}, error) {
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
	case v.Type().IsListType() || v.Type().IsTupleType():
		return ctyListToSlice(v)
	case v.Type().IsMapType() || v.Type().IsObjectType():
		return ctyMapToMap(v)
	case v.Type().IsSetType():
		return ctySetToSlice(v)
	default:
		// Instead of returning an error, let's return the string representation
		return v.GoString(), nil
	}
}

func ctySetToSlice(v cty.Value) ([]interface{}, error) {
	if !v.Type().IsSetType() {
		return nil, fmt.Errorf("not a set type")
	}

	result := make([]interface{}, 0, v.LengthInt())
	for it := v.ElementIterator(); it.Next(); {
		_, ev := it.Element()
		goValue, err := ctyValueToGo(ev)
		if err != nil {
			return nil, fmt.Errorf("error converting set element: %w", err)
		}
		result = append(result, goValue)
	}

	return result, nil
}

func ctyListToSlice(v cty.Value) ([]interface{}, error) {
	length := v.LengthInt()
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		element := v.Index(cty.NumberIntVal(int64(i)))
		goValue, err := ctyValueToGo(element)
		if err != nil {
			return nil, fmt.Errorf("error converting list element at index %d: %w", i, err)
		}
		result[i] = goValue
	}

	return result, nil
}

func ctyMapToMap(v cty.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for it := v.ElementIterator(); it.Next(); {
		key, value := it.Element()
		keyString, err := ctyValueToGo(key)
		if err != nil {
			return nil, fmt.Errorf("error converting map key: %w", err)
		}

		keyStr, ok := keyString.(string)
		if !ok {
			return nil, fmt.Errorf("map key is not a string: %v", keyString)
		}

		goValue, err := ctyValueToGo(value)
		if err != nil {
			return nil, fmt.Errorf("error converting map value for key '%s': %w", keyStr, err)
		}

		result[keyStr] = goValue
	}

	return result, nil
}

// Debug function to print cty.Value
func printCtyValue(v cty.Value) {
	fmt.Printf("Type: %s\n", v.Type().FriendlyName())
	fmt.Printf("Value: %#v\n", v)
}
