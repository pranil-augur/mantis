package helm

import (
	"fmt"

	"cuelang.org/go/cue"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// HelmTask is a task for deploying Helm charts
type HelmTask struct {
}

func NewHelmTask(val cue.Value) (hofcontext.Runner, error) {
	return &HelmTask{}, nil
}

func (t *HelmTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value
	chartConfig := v.LookupPath(cue.ParsePath("config"))

	// Extract necessary information from the CUE value
	releaseName, _ := chartConfig.LookupPath(cue.ParsePath("releaseName")).String()
	chartName, _ := chartConfig.LookupPath(cue.ParsePath("chartName")).String()
	namespace, _ := chartConfig.LookupPath(cue.ParsePath("namespace")).String()

	// Initialize Helm client
	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "", nil); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action config: %v", err)
	}

	if ctx.Preview {
		// Perform a dry-run installation
		client := action.NewInstall(actionConfig)
		client.DryRun = true
		client.ReleaseName = releaseName
		client.Namespace = namespace

		_, err := client.Run(chartName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to perform dry-run installation: %v", err)
		}

		return "Helm chart dry-run successful", nil
	} else if ctx.Apply {
		// Perform actual installation
		client := action.NewInstall(actionConfig)
		client.ReleaseName = releaseName
		client.Namespace = namespace

		_, err := client.Run(chartName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to install Helm chart: %v", err)
		}

		return "Helm chart installed successfully", nil
	} else if ctx.Destroy {
		// Uninstall the Helm release
		client := action.NewUninstall(actionConfig)

		_, err := client.Run(releaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to uninstall Helm release: %v", err)
		}

		return "Helm release uninstalled successfully", nil
	}

	return nil, fmt.Errorf("unknown command. Need to use one of preview/apply/destroy")
}
