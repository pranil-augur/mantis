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
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/command"
	"github.com/opentofu/opentofu/internal/command/cliconfig"
	"github.com/opentofu/opentofu/internal/getproviders"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/terminal"
	"github.com/opentofu/opentofu/internal/utils"
)

// TerraformDataSourceTask is a task for running a Terraform plan using a specific configuration
type TerraformDataSourceTask struct {
}

// Assuming Ui is a global variable of type cli.Ui
var Ui cli.Ui

func NewTerraformDataSourceTask(val cue.Value) (hofcontext.Runner, error) {
	return &TerraformDataSourceTask{}, nil
}

func (t *TerraformDataSourceTask) Run(ctx *hofcontext.Context) (any, error) {
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
	// Initialize commands
	commandsFactory := utils.InitCommandsWrapper(std_ctx, "", streams, config, services, providerSrc, providerDevOverrides, unmanagedProviders, scriptBytes)
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
	op, err := planCommand.RunAPI([]string{}, scriptBytes, "cue")
	if err != nil {
		return nil, fmt.Errorf("failed to execute plan command with exit status %d", err)
	}

	fmt.Println("Result status: ", op.Result.ExitStatus())
	fmt.Println(op.State)
	return op.State, nil
}
