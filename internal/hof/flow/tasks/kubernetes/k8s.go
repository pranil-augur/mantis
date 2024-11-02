/* Copyright 2024 Augur AI
/*
 * Copyright (c) 2024 Augur AI, Inc.
 * This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
 * If a copy of the MPL was not distributed with this file, you can obtain one at https://mozilla.org/MPL/2.0/.
 *
 
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
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
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

	// Nothing to do for init for Kubernetes
	if ctx.Init {
		return v, nil
	}

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
			return nil, fmt.Errorf("plan failed. Check if Kubernetes cluster is accessible: %v", err)
		}

	} else if ctx.Apply {
		// Apply the changes to the cluster
		err = client.Apply(manifest)
		if err != nil {
			return nil, fmt.Errorf("apply failed. Check if Kubernetes cluster is accessible: %v", err)
		}

	} else if ctx.Destroy {
		// Delete the specified resources
		err = client.Delete(manifest)
		if err != nil {
			return nil, fmt.Errorf("destroy failed. Check if Kubernetes cluster is accessible: %v", err)
		}
	} else if !ctx.Init { // Init has nothing to do for K8s
		return nil, fmt.Errorf("unknown command. Need to use one of plan/apply/destroy")
	}
	fmt.Println("Operation completed successfully")
	newV := v.FillPath(cue.ParsePath(mantis.MantisTaskOuts), "Resource applied successfully")
	return newV, nil
}
