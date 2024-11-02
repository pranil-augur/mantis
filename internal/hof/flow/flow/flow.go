/*
/*
 * Copyright (c) 2024 Augur AI, Inc.
 * This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
 * If a copy of the MPL was not distributed with this file, you can obtain one at https://mozilla.org/MPL/2.0/.
 *
 
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package flow

import (
	"fmt"
	"strconv"
	"strings"

	// "sync"

	"cuelang.org/go/cue"
	cueflow "cuelang.org/go/tools/flow"

	flowctx "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/flow/tasker"
	"github.com/opentofu/opentofu/internal/hof/lib/cuetils"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
)

type Flow struct {
	*hof.Node[Flow]

	Root  cue.Value
	Orig  cue.Value
	Final cue.Value

	FlowCtx *flowctx.Context
	Ctrl    *cueflow.Controller
}

func NewFlow(node *hof.Node[Flow]) *Flow {
	return &Flow{
		Node: node,
		Root: node.Value,
		Orig: node.Value,
	}
}

func OldFlow(ctx *flowctx.Context, val cue.Value) (*Flow, error) {
	p := &Flow{
		Root:    val,
		Orig:    val,
		FlowCtx: ctx,
	}
	return p, nil
}

// This is for the top-level flows
func (P *Flow) Start() error {
	err := P.run()
	// fmt.Println("Start().Err", P.Orig.Path(), err)
	return err
}

func (P *Flow) run() error {
	if P.Node == nil {
		node, err := hof.ParseHof[Flow](P.Orig)
		if err != nil {
			return err
		}
		if node == nil {
			return fmt.Errorf("Root flow value is not a flow, has nil #hof node")
		}
		P.Node = node
	}

	// fmt.Println("FLOW.run:", P.FlowCtx.RootValue.Path(), P.Root.Path())
	// root := P.FlowCtx.RootValue
	root := P.Root
	// Setup the flow Config
	cfg := &cueflow.Config{
		// InferTasks:      true,
		IgnoreConcrete:  true,
		FindHiddenTasks: true,
		UpdateFunc: func(c *cueflow.Controller, t *cueflow.Task) error {
			//if t != nil {
			//  fmt.Println("Flow.Update()", t.Index(), t.Path())
			//} else {
			//  fmt.Println("Flow.Update()", "nil task")
			//}
			if t != nil {
				v := t.Value()

				node, err := hof.ParseHof[any](v)
				if err != nil {
					return err
				}
				if node == nil {
					panic("we should have found a node to even get here")
				}

				if node.Hof.Flow.Task == "" {
					return nil
				}

				if node.Hof.Flow.Print.Level > 0 && !node.Hof.Flow.Print.Before {
					pv := v.LookupPath(cue.ParsePath(node.Hof.Flow.Print.Path))
					if node.Hof.Path == "" {
						fmt.Printf("%s", node.Hof.Flow.Print.Path)
					} else if node.Hof.Flow.Print.Path == "" {
						fmt.Printf("%s", node.Hof.Path)
					} else {
						fmt.Printf("%s.%s", node.Hof.Path, node.Hof.Flow.Print.Path)
					}
					fmt.Printf(": %v\n", pv)
				}
			}
			return nil
		},
	}

	// This is for flows down from the root val
	// This is needed because nested flows (like IRC / API handler)
	// ... break if this check is not performed
	// ... and we blindly set the RootPath the value Path
	if P.Orig != P.Root {
		cfg.Root = P.Orig.Path()
	}

	// copy orig for good measure
	// This is helpful for when
	v := P.Orig.Context().CompileString("{...}")
	u := v.Unify(root)

	// create the workflow which will build the task graph
	P.Ctrl = cueflow.New(cfg, u, tasker.NewTasker(P.FlowCtx))

	if P.FlowCtx.Plan || P.FlowCtx.Gist {
		P.createAndPrintMantisPlan()
		P.createAndPrintMantisGraph()
	}

	// fmt.Println("Flow.run() start")
	err := P.Ctrl.Run(P.FlowCtx.GoContext)

	//print error from ctx.FlowErrors and ctx.FlowWarnings
	if len(P.FlowCtx.FlowErrors) > 0 || len(P.FlowCtx.FlowWarnings) > 0 {
		for _, err := range P.FlowCtx.FlowErrors {
			cuetils.PrintWarningOrError(false, err)
		}
		for _, warning := range P.FlowCtx.FlowWarnings {
			cuetils.PrintWarningOrError(true, warning)
		}
	}

	// fmt.Println("Flow.run() end", err)

	// fmt.Println("flow(end):", P.path, P.rpath)
	P.Final = P.Ctrl.Value()
	if err != nil {
		s := cuetils.CueErrorToString(err)
		// fmt.Println("Flow ERR in?", P.Orig.Path(), s)

		//fmt.Println(P)
		return fmt.Errorf("Error in %s | %s: %s", P.Hof.Metadata.Name, P.Orig.Path(), s)
	}
	// fmt.Println("NOT HERE", P.Orig.Path())

	return nil
}

func (P *Flow) createAndPrintMantisGraph() (map[string]interface{}, error) {
	tasks := P.Ctrl.Tasks()
	for _, t := range tasks {
		fmt.Println("Task:", t.Path())
		for _, dep := range t.Dependencies() {
			fmt.Println("  Depends on:", dep.Path())
		}
	}

	return nil, nil
}

func (P *Flow) createAndPrintMantisPlan() (map[string]interface{}, error) {
	tfResources := make(map[string]interface{})
	v := P.Root

	// Recursively scan the CUE value tree for resources
	err := scanForLabels(v, cue.Path{}, tfResources, "resource", "module")
	if err != nil {
		return nil, fmt.Errorf("error scanning for resources: %v", err)
	}

	// Scan for Kubernetes resources
	kubernetesLabels := make([]string, 0, len(mantis.MantisKubernetesResourceNames))
	for _, resourceName := range mantis.MantisKubernetesResourceNames {
		kubernetesLabels = append(kubernetesLabels, resourceName)
	}
	err = scanForLabels(v, cue.Path{}, tfResources, kubernetesLabels...)
	if err != nil {
		return nil, fmt.Errorf("error scanning for Kubernetes resources: %v", err)
	}

	if len(tfResources) == 0 {
		return nil, fmt.Errorf("no resources found in the CUE configuration")
	}

	if len(tfResources) > 0 {
		fmt.Println("Quick Resource Summary:")
		fmt.Println("---------------------------")
		for resourceType, resources := range tfResources {
			count := 0
			switch r := resources.(type) {
			case map[string]interface{}:
				count = len(r)
			case []interface{}:
				count = len(r)
			}
			fmt.Printf("%s (%d)\n", resourceType, count)
			switch r := resources.(type) {
			case map[string]interface{}:
				printResourceTree(r, 1)
			case []interface{}:
				for _, resource := range r {
					if m, ok := resource.(map[string]interface{}); ok {
						printResourceTree(m, 1)
					}
				}
			}
		}
		fmt.Println("---------------------------")
	} else {
		fmt.Println("No resources found in the plan.")
	}
	return tfResources, nil
}

func scanForLabels(v cue.Value, path cue.Path, resources map[string]interface{}, labels ...string) error {
	switch v.Kind() {
	case cue.StructKind:
		iter, err := v.Fields()
		if err != nil {
			return fmt.Errorf("error iterating over fields: %v", err)
		}

		for iter.Next() {
			label := iter.Label()
			newPath := cue.ParsePath(path.String() + "." + label)
			if contains(labels, label) {
				err := extractLabeledItem(iter.Value(), newPath, resources, label)
				if err != nil {
					return err
				}
			} else {
				err := scanForLabels(iter.Value(), newPath, resources, labels...)
				if err != nil {
					return err
				}
			}
		}

	case cue.ListKind:
		list, err := v.List()
		if err != nil {
			return fmt.Errorf("error getting list: %v", err)
		}

		for i := 0; list.Next(); i++ {
			newPath := cue.ParsePath(path.String() + "." + strconv.Itoa(i))
			err := scanForLabels(list.Value(), newPath, resources, labels...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractLabeledItem(v cue.Value, path cue.Path, resources map[string]interface{}, label string) error {
	iter, err := v.Fields()
	if err != nil {
		return fmt.Errorf("error iterating over %s fields: %v", label, err)
	}

	itemMap, ok := resources[label].(map[string]interface{})
	if !ok {
		itemMap = make(map[string]interface{})
		resources[label] = itemMap
	}

	for iter.Next() {
		itemType := iter.Label()
		itemValue := iter.Value()

		if label == "module" {
			// For modules, we don't have an additional nesting level
			configMap, err := extractConfigMap(itemValue)
			if err != nil {
				return fmt.Errorf("error extracting config for %s %s: %v", label, itemType, err)
			}
			itemMap[itemType] = configMap
		} else {
			// For resources, we keep the existing structure
			itemIter, err := itemValue.Fields()
			if err != nil {
				return fmt.Errorf("error iterating over %s %s: %v", label, itemType, err)
			}

			if _, ok := itemMap[itemType]; !ok {
				itemMap[itemType] = make(map[string]interface{})
			}

			for itemIter.Next() {
				itemName := itemIter.Label()
				itemConfig := itemIter.Value()

				configMap, err := extractConfigMap(itemConfig)
				if err != nil {
					return fmt.Errorf("error extracting config for %s %s.%s: %v", label, itemType, itemName, err)
				}

				itemMap[itemType].(map[string]interface{})[itemName] = configMap
			}
		}
	}

	return nil
}

func extractConfigMap(v cue.Value) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	switch v.Kind() {
	case cue.StructKind:
		iter, err := v.Fields()
		if err != nil {
			return nil, fmt.Errorf("error iterating over config fields: %v", err)
		}

		for iter.Next() {
			field := iter.Label()
			value := iter.Value()

			var goValue interface{}
			var err error
			goValue, err = value.String()
			if err != nil {
				goValue = fmt.Sprintf("%v", value)
			}

			configMap[field] = goValue
		}
	default:
		// For non-struct types (like strings), just return the value directly
		var goValue interface{}
		err := v.Decode(&goValue)
		if err != nil {
			return nil, fmt.Errorf("error decoding value: %v", err)
		}
		return map[string]interface{}{"value": goValue}, nil
	}

	return configMap, nil
}

func printResourceTree(resource map[string]interface{}, depth int) {
	indent := strings.Repeat("  ", depth)
	for name, details := range resource {
		fmt.Printf("%s├─ %s", indent, name)
		if detailMap, ok := details.(map[string]interface{}); ok {
			if len(detailMap) > 0 {
				fmt.Println()
				//printResourceDetails(detailMap, depth+1)
			} else {
				fmt.Println()
			}
		} else {
			fmt.Println()
		}
	}
}
