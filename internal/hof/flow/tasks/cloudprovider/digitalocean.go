/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package cloudprovider

import (
	"context"
	"fmt"
	"path"

	"cuelang.org/go/cue"
	"github.com/digitalocean/godo"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
)

type DigitalOceanTask struct{}

func NewDigitalOceanTask(val cue.Value) (hofcontext.Runner, error) {
	return &DigitalOceanTask{}, nil
}

func (t *DigitalOceanTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value

	// Extract the configuration from the CUE value
	configValue := v.LookupPath(cue.ParsePath("config"))

	// Initialize DigitalOcean client
	client, err := newDigitalOceanClient(configValue)
	if err != nil {
		return nil, fmt.Errorf("failed to create DigitalOcean client: %v", err)
	}

	// Query resources based on the configuration
	resources, err := queryResources(client, configValue)
	if err != nil {
		return nil, fmt.Errorf("failed to query DigitalOcean resources: %v", err)
	}

	// Update the CUE value with the query results
	newV := v.FillPath(cue.ParsePath(mantis.MantisTaskOuts), resources)
	return newV, nil
}

func newDigitalOceanClient(config cue.Value) (*godo.Client, error) {
	token, err := config.LookupPath(cue.ParsePath("token")).String()
	if err != nil {
		return nil, fmt.Errorf("failed to get DigitalOcean API token: %v", err)
	}

	return godo.NewFromToken(token), nil
}

func queryResources(client *godo.Client, config cue.Value) (map[string]interface{}, error) {
	resourceTypeStr, err := config.LookupPath(cue.ParsePath("resourceType")).String()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource type: %v", err)
	}

	resourceType := mantis.DOResource(resourceTypeStr)

	switch resourceType {
	case mantis.Droplet:
		return queryDroplets(client, config)
	case mantis.Volume:
		return queryVolumes(client, config)
	// Add more resource types as needed
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func queryDroplets(client *godo.Client, config cue.Value) (map[string]interface{}, error) {
	ctx := context.Background()
	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Get filters from config
	filters := struct {
		Name   string `json:"name"`
		Region string `json:"region"`
		Tag    string `json:"tag"`
		Status string `json:"status"`
		Size   string `json:"size"`
	}{}

	if filtersValue := config.LookupPath(cue.ParsePath("filters")); filtersValue.Exists() {
		if err := filtersValue.Decode(&filters); err != nil {
			return nil, fmt.Errorf("failed to decode filters: %v", err)
		}
	}

	var filteredDroplets []godo.Droplet
	for _, droplet := range droplets {
		if !matchesDropletFilters(droplet, filters) {
			continue
		}
		filteredDroplets = append(filteredDroplets, droplet)
	}

	return map[string]interface{}{"droplets": filteredDroplets}, nil
}

func matchesDropletFilters(droplet godo.Droplet, filters struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Tag    string `json:"tag"`
	Status string `json:"status"`
	Size   string `json:"size"`
}) bool {
	// Check name filter (supports wildcard matching)
	if filters.Name != "" {
		matched, _ := path.Match(filters.Name, droplet.Name)
		if !matched {
			return false
		}
	}

	// Check region filter
	if filters.Region != "" && droplet.Region.Slug != filters.Region {
		return false
	}

	// Check tag filter
	if filters.Tag != "" && !containsTag(droplet.Tags, filters.Tag) {
		return false
	}

	// Check status filter
	if filters.Status != "" && droplet.Status != filters.Status {
		return false
	}

	// Check size filter
	if filters.Size != "" && droplet.Size.Slug != filters.Size {
		return false
	}

	return true
}

func queryVolumes(client *godo.Client, config cue.Value) (map[string]interface{}, error) {
	ctx := context.Background()
	volumes, _, err := client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{})
	if err != nil {
		return nil, err
	}

	// Get filters from config
	filters := struct {
		Name   string `json:"name"`
		Region string `json:"region"`
		Size   string `json:"size"`
		Tag    string `json:"tag"`
	}{}

	if filtersValue := config.LookupPath(cue.ParsePath("filters")); filtersValue.Exists() {
		if err := filtersValue.Decode(&filters); err != nil {
			return nil, fmt.Errorf("failed to decode filters: %v", err)
		}
	}

	var filteredVolumes []godo.Volume
	for _, volume := range volumes {
		if !matchesVolumeFilters(volume, filters) {
			continue
		}
		filteredVolumes = append(filteredVolumes, volume)
	}

	return map[string]interface{}{"volumes": filteredVolumes}, nil
}

func matchesVolumeFilters(volume godo.Volume, filters struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Size   string `json:"size"`
	Tag    string `json:"tag"`
}) bool {
	// Check name filter (supports wildcard matching)
	if filters.Name != "" {
		matched, _ := path.Match(filters.Name, volume.Name)
		if !matched {
			return false
		}
	}

	// Check region filter
	if filters.Region != "" && volume.Region.Slug != filters.Region {
		return false
	}

	// Check size filter (convert to string for comparison)
	if filters.Size != "" {
		volumeSize := fmt.Sprintf("%d", volume.SizeGigaBytes)
		if volumeSize != filters.Size {
			return false
		}
	}

	// Check tag filter
	if filters.Tag != "" && !containsTag(volume.Tags, filters.Tag) {
		return false
	}

	return true
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
