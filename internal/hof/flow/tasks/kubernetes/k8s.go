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
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type K8sTask struct{}

func NewK8sTask(val cue.Value) (hofcontext.Runner, error) {
	return &K8sTask{}, nil
}

func (t *K8sTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value
	k8sConfig := v.LookupPath(cue.ParsePath("config"))

	// Extract the manifest data from the CUE value
	manifests, err := k8sConfig.LookupPath(cue.ParsePath("manifests")).Bytes()
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
		err = client.Plan(manifests)
		if err != nil {
			return nil, fmt.Errorf("plan failed: %v", err)
		}

		return "Kubernetes resources dry-run successful", nil

	} else if ctx.Apply {
		// Apply the changes to the cluster
		err = client.Apply(manifests)
		if err != nil {
			return nil, fmt.Errorf("apply failed: %v", err)
		}

		return "Kubernetes resources applied successfully", nil

	} else if ctx.Destroy {
		// Delete the specified resources
		err = client.Delete(manifests)
		if err != nil {
			return nil, fmt.Errorf("destroy failed: %v", err)
		}

		return "Kubernetes resources deleted successfully", nil
	}

	return nil, fmt.Errorf("unknown command. Need to use one of plan/apply/destroy")
}
