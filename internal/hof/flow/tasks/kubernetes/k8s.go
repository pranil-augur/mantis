/* Copyright 2024 Augur AI
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 */

package kubernetes

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/pkg/encoding/yaml"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type K8sTask struct{}

func NewK8sTask(val cue.Value) (hofcontext.Runner, error) {
	return &K8sTask{}, nil
}

func (t *K8sTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value
	// Extract the manifest data from the CUE value
	configValue := v.LookupPath(cue.ParsePath("config"))
	manifest, err := yaml.Marshal(configValue)

	if err != nil {
		return nil, fmt.Errorf("failed to extract manifests from CUE: %v", err)
	}

	// Initialize Kubernetes client
	client, err := NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	if ctx.Plan {
		// Perform a dry-run to simulate changes
		err = client.Plan(string(manifest))
		if err != nil {
			return nil, fmt.Errorf("plan failed: %v", err)
		}

	} else if ctx.Apply {
		// Apply the changes to the cluster
		err = client.Apply(manifest)
		if err != nil {
			return nil, fmt.Errorf("apply failed: %v", err)
		}

	} else if ctx.Destroy {
		// Delete the specified resources
		err = client.Delete(manifest)
		if err != nil {
			return nil, fmt.Errorf("destroy failed: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unknown command. Need to use one of plan/apply/destroy")
	}
	fmt.Println("Operation completed successfully")
	newV := v.FillPath(cue.ParsePath(hof.MantisTaskOuts), "Resource applied successfully")
	return newV, nil
}
