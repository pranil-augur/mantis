package cueform

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/opentofu/opentofu/internal/addrs"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

// TerraformDataSourceTask is a task for fetching data using a Terraform data source API
type TerraformDataSourceTask struct {
	// BaseTask
	Provider provider.Provider

	DataSourceName string
}

type DataSourceConfig struct {
	ProviderType   string
	Config         map[string]interface{}
	DataSourceName string
}

func NewTerraformDataSourceTask(val cue.Value, config DataSourceConfig) (hofcontext.Runner, error) {
	provider, err := InitializeProvider(config.ProviderType, config.Config)
	if err != nil {
		return nil, err
	}
	return &TerraformDataSourceTask{
		Provider:       provider,
		DataSourceName: config.DataSourceName,
	}, nil
}

func InitializeProvider(providerType string, config map[string]interface{}) (provider.Provider, error) {
	switch providerType {
	case "aws":
		// Assuming AWS is a default provider under the hashicorp namespace
		awsProvider := addrs.NewProvider(addrs.DefaultProviderRegistryHost, "hashicorp", "aws")
		// return NewAWSProvider(config, awsProvider)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

func (t *TerraformDataSourceTask) Run(ctx *hofcontext.Context) (any, error) {
	fmt.Println("Fetching data from Terraform data source:", t.DataSourceName)

	tfCtx := context.Background()
	dataSourceFuncs := t.Provider.DataSources(tfCtx)
	var selectedDataSource datasource.DataSource
	for _, dsFunc := range dataSourceFuncs {
		ds := dsFunc()
		if ds.Name() == t.DataSourceName {
			selectedDataSource = ds
			break
		}
	}
	if selectedDataSource == nil {
		return nil, fmt.Errorf("data source %s not found", t.DataSourceName)
	}

	readResp := datasource.ReadResponse{}
	selectedDataSource.Read(tfCtx, &datasource.ReadRequest{}, &readResp)
	if readResp.Diagnostics.HasError() {
		return nil, fmt.Errorf("error reading from data source: %v", readResp.Diagnostics)
	}

	fetchedData := readResp.State.GetAttributeValues()
	return fetchedData, nil
}

// GetDataSourcesForProvider fetches all data sources available for a given provider.
func GetDataSourcesForProvider(providerName string) ([]string, error) {
	// Initialize the provider based on the provider name
	provider, err := InitializeProvider(providerName, nil)
	if err != nil {
		return nil, err
	}

	// Get the data sources from the provider
	tfCtx := context.Background()
	dataSourceFuncs := provider.DataSources(tfCtx)
	dataSourceNames := make([]string, 0, len(dataSourceFuncs))
	for _, dsFunc := range dataSourceFuncs {
		ds := dsFunc()
		dataSourceNames = append(dataSourceNames, ds.Name())
	}

	return dataSourceNames, nil
}
