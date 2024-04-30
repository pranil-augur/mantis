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
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/command"
	"github.com/opentofu/opentofu/internal/command/cliconfig"
	"github.com/opentofu/opentofu/internal/command/views"
	"github.com/opentofu/opentofu/internal/getproviders"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/terminal"
)

// TerraformDataSourceTask is a task for running a Terraform plan using a specific configuration
type TerraformDataSourceTask struct {
	Provider       provider.Provider
	ConfigFilePath string
}

// Assuming Ui is a global variable of type cli.Ui
var Ui cli.Ui

func NewTerraformDataSourceTask(val cue.Value, configFilePath string) (hofcontext.Runner, error) {
	return &TerraformDataSourceTask{
		Provider:       provider,
		ConfigFilePath: configFilePath,
	}, nil
}

func (t *TerraformDataSourceTask) Run(ctx *hofcontext.Context) (any, error) {
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

	// Initialize commands
	initCommands(ctx.Context, "", streams, config, services, providerSrc, providerDevOverrides, unmanagedProviders)

	originalWd, err := os.Getwd()
	wd := workingDir(originalWd, os.Getenv("TF_DATA_DIR"))

	// Setup the environment for running the PlanCommand
	meta := command.Meta{
		WorkingDir: wd,
		Streams:    streams,
		View:       views.NewView(streams),
		Ui:         Ui,
	}

	// Initialize the PlanCommand with the meta configuration
	planCommand := &command.PlanCommand{
		Meta: meta,
	}

	// Execute the PlanCommand with the configuration file path
	exitStatus := planCommand.Run([]string{t.ConfigFilePath})
	if exitStatus != 0 {
		return nil, fmt.Errorf("failed to execute plan command with exit status %d", exitStatus)
	}

	fmt.Println("Plan command executed successfully.")
	return "Plan executed successfully", nil
}
