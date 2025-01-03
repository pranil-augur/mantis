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
	"fmt"
	"regexp"
	"time"

	"github.com/opentofu/opentofu/internal/hof/lib/datamodel"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type DatamodelEnricher func(*Runtime, *datamodel.Datamodel) error

func (R *Runtime) EnrichDatamodels(datamodels []string, enrich DatamodelEnricher) error {
	start := time.Now()
	defer func() {
		end := time.Now()
		R.Stats.Add("enrich/data", end.Sub(start))
	}()

	if R.Flags.Verbosity > 1 {
		fmt.Println("Runtime.EnrichDatamodels: ", datamodels)
		for _, node := range R.Nodes {
			node.Print()
		}
	}

	keep := func(hn *hof.Node[any]) bool {

		// We only want datamodels at the root of the Node Tree
		// Otherwise the DM is nested for usage elsewhere
		if hn.Parent != nil {
			return false
		}

		// filter by name
		if len(datamodels) > 0 {
			for _, d := range datamodels {
				match, err := regexp.MatchString(d, hn.Hof.Metadata.Name)
				if err != nil {
					fmt.Println("error:", err)
					return false
				}
				if match {
					return true
				}
			}
			return false
		}

		// filter by time

		// filter by version?

		// default to true, should include everything when no checks are needed
		return true
	}

	// Find only the datamodel nodes, these are all root nodes (in theory)
	// TODO, dedup any references
	dms := []*datamodel.Datamodel{}
	for _, node := range R.Nodes {
		// check for DM root
		if node.Hof.Datamodel.Root {
			if !keep(node) {
				continue
			}

			upgrade := func(n *hof.Node[datamodel.Value]) *datamodel.Value {
				v := new(datamodel.Value)
				v.Node = n
				v.Snapshot = new(datamodel.Snapshot)
				return v
			}

			dm := hof.Upgrade[any, datamodel.Value](node, upgrade, nil)
			// we'd like this line in upgrade, but...
			// how do we make T a Node[T] type (or ensure that it has a hof)
			// u.T.Hof = u.Hof
			dms = append(dms, &datamodel.Datamodel{Node: dm})
		}
	}

	R.Datamodels = dms

	// filter datamodel if flag set?
	// which flags do we handle here
	//   vs in the various commands?

	// load history
	// calc diffs? (or load)

	for _, dm := range R.Datamodels {
		err := enrich(R, dm)
		if err != nil {
			return err
		}
	}


	return nil
}
