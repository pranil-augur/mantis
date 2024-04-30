/*
 * Augur AI Proprietary
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
	"time"

	"github.com/opentofu/opentofu/internal/hof/lib/gen"
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type GeneratorEnricher func(*Runtime, *gen.Generator) error

func (R *Runtime) EnrichGenerators(generators []string, enrich GeneratorEnricher) error {
	start := time.Now()
	defer func() {
		end := time.Now()
		R.Stats.Add("enrich/gen", end.Sub(start))
	}()

	if R.Flags.Verbosity > 1 {
		fmt.Println("Runtime.EnrichGenerators: ", generators)
		for _, node := range R.Nodes {
			node.Print()
		}
	}

	// Find only the generator nodes
	// TODO, dedup any references
	gens := []*gen.Generator{}
	for _, node := range R.Nodes {
		// check for Gen root
		if node.Hof.Gen.Root {
			if !keepFilter(node, generators) {
				continue
			}
			upgrade := func(n *hof.Node[gen.Generator]) *gen.Generator {
				v := gen.NewGenerator(n)
				v.RootModuleName = R.BuildInstances[0].Module
				return v
			}
			u := hof.Upgrade[any, gen.Generator](node, upgrade, nil)
			// we'd like this line in upgrade, but...
			// how do we make T a Node[T] type (or ensure that it has a hof)
			// u.T.Hof = u.Hof
			gen := u.T
			gen.Node = u
			gens = append(gens, gen)
		}
	}

	R.Generators = gens

	// what do we do to enrich a generator?
	// load & validate?
	// add datamodel history to input data?

	for _, gen := range R.Generators {
		err := enrich(R, gen)
		if err != nil {
			return err
		}
	}

	return nil
}
