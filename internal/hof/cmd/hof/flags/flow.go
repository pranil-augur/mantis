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

package flags

import (
	"github.com/spf13/pflag"
)

var _ *pflag.FlagSet

var FlowFlagSet *pflag.FlagSet

type FlowPflagpole struct {
	Flow     []string
	Bulk     string
	Parallel int
	Progress bool
}

func SetupFlowPflags(fset *pflag.FlagSet, fpole *FlowPflagpole) {
	// pflags

	fset.StringArrayVarP(&(fpole.Flow), "flow", "F", nil, "flow labels to match and run")
	fset.StringVarP(&(fpole.Bulk), "bulk", "B", "", "exprs for inputs to run workflow in bulk mode")
	fset.IntVarP(&(fpole.Parallel), "parallel", "P", 1, "global flow parallelism")
	fset.BoolVarP(&(fpole.Progress), "progress", "", false, "print task progress as it happens")
}

var FlowPflags FlowPflagpole

func init() {
	FlowFlagSet = pflag.NewFlagSet("Flow", pflag.ContinueOnError)

	SetupFlowPflags(FlowFlagSet, &FlowPflags)

}
