package opentf

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/opentofu/opentofu/internal/addrs"
	backendInit "github.com/opentofu/opentofu/internal/backend/init"
	"github.com/opentofu/opentofu/internal/command"
	"github.com/opentofu/opentofu/internal/command/cliconfig"
	"github.com/opentofu/opentofu/internal/configs"
	"github.com/opentofu/opentofu/internal/getproviders"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/terminal"
	"github.com/opentofu/opentofu/internal/utils"
)

// TFTask is a task for running a Terraform plan using a specific configuration
type TFProvidersTask struct {
}

func NewTFProvidersTask(val cue.Value) (hofcontext.Runner, error) {
	return &TFProvidersTask{}, nil
}

func (t *TFProvidersTask) Run(ctx *hofcontext.Context) (any, error) {
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

	// Initialize the backends.
	backendInit.Init(services)
	var std_ctx context.Context

	taskPath := ctx.BaseTask.ID
	configDetails := &configs.MantisConfig{
		Identifier: taskPath,
		Content:    scriptBytes,
		Format:     "json",
	}
	// Initialize commands
	commandsFactory := utils.InitCommandsWrapper(std_ctx, "", streams, config, services, providerSrc, providerDevOverrides, unmanagedProviders, configDetails)
	if ctx.Init {
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

		retval := initCommand.Run([]string{"-reconfigure"})
		if retval < 0 {
			return nil, fmt.Errorf("error Initializing")
		}
	}
	return nil, nil
}
