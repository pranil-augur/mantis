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

package runtime

import (
	"time"

	"github.com/opentofu/opentofu/internal/hof/flow/flow"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type FlowEnricher func(*Runtime, *flow.Flow) error

func (R *Runtime) EnrichFlows(flows []string, enrich FlowEnricher) error {
	// fmt.Println("Flow.Runtime.Enrich", flows)
	start := time.Now()
	defer func() {
		end := time.Now()
		R.Stats.Add("enrich/flow", end.Sub(start))
	}()

	if R.Flags.Verbosity > 1 {
		//fmt.Println("Runtime.Flow: ", flows)
		//for _, node := range R.Nodes {
		//  node.Print()
		//}
	}

	// Find only the datamodel nodes
	// TODO, dedup any references
	fs := []*flow.Flow{}
	for _, node := range R.Nodes {
		// check for Chat root
		if node.Hof.Flow.Root {
			// fmt.Println("flow.root:", node.Hof.Path)
			if !keepFilter(node, flows) {
				continue
			}
			upgrade := func(n *hof.Node[flow.Flow]) *flow.Flow {
				v := flow.NewFlow(n)
				return v
			}
			u := hof.Upgrade[any, flow.Flow](node, upgrade, nil)
			// we'd like this line in upgrade, but...
			// how do we make T a Node[T] type (or ensure that it has a hof)
			// u.T.Hof = u.Hof
			f := u.T
			f.Node = u
			fs = append(fs, f)
		}
	}

	R.Workflows = fs

	for _, c := range R.Workflows {
		err := enrich(R, c)
		if err != nil {
			return err
		}
	}


	return nil
}
