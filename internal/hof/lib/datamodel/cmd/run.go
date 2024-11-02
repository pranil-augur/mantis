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

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/datamodel"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func Run(cmd string, args []string, rflags flags.RootPflagpole, dflags flags.DatamodelPflagpole) error {
	// fmt.Printf("lib/datamodel.Run.%s %v %v %v\n", cmd, args, rflags, dflags)

	R, err := runtime.New(args, rflags)
	if err != nil {
		return err
	}

	err = R.Load()
	if err != nil {
		return err
	}

	err = R.EnrichDatamodels(dflags.Datamodels, EnrichDatamodel)
	if err != nil {
		return err
	}

	// fmt.Println("R.dms:", len(R.Datamodels))

	// Now, with our datamodles at hand, run the command
	switch cmd {
	case "list":
		err = list(R, dflags)

	case "tree":
		err = tree(R, dflags)

	case "checkpoint":
		err = checkpoint(R, dflags, flags.Datamodel__CheckpointFlags)

	case "diff":
		err = diff(R, dflags)

	case "log":
		err = log(R, dflags)

	default:
		err = fmt.Errorf("%s command not implemented yet", cmd)
	}

	return err
}

func EnrichDatamodel(R *runtime.Runtime, dm *datamodel.Datamodel) error {
	err := dm.LoadHistory()
	if err != nil {
		return err
	}
	err = dm.CalcDiffs()
	if err != nil {
		return err
	}

	// fmt.Println("enriched:", dm.Hof.Path, dm.T.Value)
	// fmt.Println(R.Value)

	return nil
}
