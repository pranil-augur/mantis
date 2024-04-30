package cueform

import (
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/mitchellh/cli"
	"github.com/opentofu/opentofu/internal/command"
	"github.com/opentofu/opentofu/internal/command/views"
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
	wd := workingDir(originalWorkingDir, os.Getenv("TF_DATA_DIR"))
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	streams := terminal.DefaultStreams() // Assuming no error is returned

	meta := command.Meta{
		WorkingDir: workingDir,
		Streams:    streams,
		View:       views.NewView(streams),
		Ui:         Ui, // Use the global Ui variable
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
