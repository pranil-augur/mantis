/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package cueform

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
	script := v.LookupPath(cue.ParsePath("script"))
	jsonScript, err := script.MarshalJSON()
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
	if ctx.Preview {
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
		v.FillPath(cue.ParsePath("out"), parsedVariables)
		// Attempt to fill the path with the new value
		newV := v.FillPath(cue.ParsePath("out"), parsedVariables)

		return newV, nil
	} else if ctx.Apply {
		// Retrieve the 'plan' command from the commandsFactory using the appropriate key
		applyCommandFactory, exists := commandsFactory["apply"]
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
		terraform := value.LookupPath(cue.ParsePath("terraform"))
		if terraform.Exists() {
			fmt.Println("Running init with " + scriptStr)
		} else {
			fmt.Println("Skipping  init, no terraform" + scriptStr)
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
			return nil, fmt.Errorf("Error Initializing")
		}
	} else {
		return nil, fmt.Errorf("Unknown command: ")
	}
	return nil, nil
}
