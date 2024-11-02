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

package cuecmd

import (
	"fmt"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/cuetils"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func Def(args []string, rflags flags.RootPflagpole, cflags flags.DefFlagpole) error {

	start := time.Now()
	R, err := runtime.New(args, rflags)

	defer func() {
		if R.Flags.Stats {
			fmt.Println(R.Stats)
			end := time.Now()
			fmt.Printf("\nTotal Elapsed Time: %s\n\n", end.Sub(start))
		}
	}()

	if err != nil {
		return err
	}

	err = R.Load()
	if err != nil {
		return cuetils.ExpandCueError(err)
	}

	// build options
	opts := []cue.Option{
		cue.Docs(cflags.Comments),
		cue.Attributes(cflags.Attributes),
		cue.InlineImports(cflags.InlineImports),
		cue.Concrete(false),
	}

	fopts := []format.Option{}
	if cflags.Simplify {
		fopts = append(fopts, format.Simplify())
	}

	bi := R.BuildInstances[0]
	if R.Flags.Verbosity > 1 {
		fmt.Println("ID:", bi.ID(), bi.PkgName, bi.Module)
	}
	pkg := bi.PkgName
	if bi.Module == "" {
		pkg = bi.ID()
	}
	err = writeOutput(R.Value, pkg, opts, fopts, cflags.Out, cflags.Outfile, cflags.Expression, rflags.Schema, false, false, true)
	if err != nil {
		return err
	}

	return nil
}
