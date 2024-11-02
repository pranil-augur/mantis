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

package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/cuetils"
	"github.com/opentofu/opentofu/internal/hof/lib/datamodel"
	"github.com/opentofu/opentofu/internal/hof/lib/gen"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

// gen.Runtime extends the common runtime.Runtime
type Runtime struct {
	*runtime.Runtime

	// Setup options
	GenFlags     flags.GenFlagpole
	Diff3FlagSet bool // this is so we can set it to true without and explicit "true"
}

func NewGenRuntime(RT *runtime.Runtime, gflags flags.GenFlagpole) (*Runtime) {
	return &Runtime{
		Runtime:  RT,
		GenFlags: gflags,
	}

}

func prepRuntime(args []string, rflags flags.RootPflagpole, gflags flags.GenFlagpole) (*Runtime, error) {

	// create our core runtime
	r, err := runtime.New(args, rflags)
	if err != nil {
		return nil, err
	}
	// upgrade to a generator runtime
	R := NewGenRuntime(r, gflags)

	// log cue dirs
	if R.Flags.Verbosity > 1 {
		fmt.Println("CueDirs:", R.CueModuleRoot, R.WorkingDir, R.CwdToRoot)
	}

	// First time load (not-fast)
	err = R.Reload(false)
	if err != nil {
		cuetils.PrintCueError(err)
		return R, fmt.Errorf("while loading generators")
	}

	// fmt.Printf("%# +v\n", R.Value)

	if len(R.Generators) == 0 {
		return R, fmt.Errorf("no generators found")
	}

	// run pre-flows here?

	return R, nil
}


func (R *Runtime) Clear() {
	R.Datamodels = make([]*datamodel.Datamodel, 0, len(R.Datamodels))
	R.Generators = make([]*gen.Generator, 0, len(R.Generators))
}

func (R *Runtime) WriteOutput() []error {
	var errs []error
	if R.Flags.Verbosity > 0 {
		fmt.Println("Writing output")
	}

	for _, G := range R.Generators {
		gerrs := G.Write(R.OutputDir(R.GenFlags.Outdir), R.ShadowDir())
		errs = append(errs, gerrs...)
	}

	return errs
}

const SHADOW_DIR = ".hof/shadow/"

// ShadowDir returns the absolute path to shadow dir for this runtime.
// It accounts for module root and relative directories.
func (R *Runtime) ShadowDir() string {
	return filepath.Join(R.CueModuleRoot, SHADOW_DIR, R.RootToCwd, R.GenFlags.Outdir)
}

func (R *Runtime) RunGenerators() []error {
	start := time.Now()
	defer func() {
		end := time.Now()
		R.Stats.Add("gen/run", end.Sub(start))
	}()

	var errs []error

	// Load shadow, can this be done in parallel with the last step?
	// Don't do in parallel yet, Cue can be slow and hungry for memory
	// CUE is not concurrency safe yet, even if, this doesn't take that long anyway
	for _, G := range R.Generators {
		gerrs := R.RunGenerator(G)
		if len(gerrs) > 0 {
			errs = append(errs, gerrs...)
		}
	}

	return errs
}

func (R *Runtime) RunGenerator(G *gen.Generator) (errs []error) {
	if G.Disabled {
		return
	}

	outputDir := filepath.Join(R.OutputDir(R.GenFlags.Outdir), G.OutputPath())
	shadowDir := filepath.Join(R.ShadowDir(), G.ShadowPath())

	// late load shadow, only if we are going to generate
	err := G.LoadShadow(shadowDir)
	if err != nil {
		return []error{err}
	}

	// fmt.Println(G.CueValue)

	// run this generator
	errsG := G.GenerateFiles(outputDir)
	if len(errsG) > 0 {
		errs = append(errs, errsG...)
		return errs
	}

	// run any subgenerators
	for _, sg := range G.Generators {
		// make sure
		sg.UseDiff3 = G.UseDiff3
		sgerrs := R.RunGenerator(sg)
		if len(sgerrs) > 0 {
			errs = append(errs, sgerrs...)
		}
	}

	return errs
}

func (R *Runtime) PrintStats() {
	// find gens which ran
	gens := []string{}
	for _, G := range R.Generators {
		if !G.Disabled {
			gens = append(gens, G.Name)
		}
	}

	fmt.Printf("\nHof: %s\n==========================\n", "Runtime")
	fmt.Println("\nGens:", gens)
	fmt.Println(R.Stats)

	for _, G := range R.Generators {
		if G.Disabled {
			continue
		}

		G.Stats.CalcTotals(G)
		fmt.Printf("\nGen: %s\n==========================\n", G.Name)
		fmt.Println(G.Stats)
	}
}

