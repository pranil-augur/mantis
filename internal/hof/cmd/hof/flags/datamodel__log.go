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

var Datamodel__LogFlagSet *pflag.FlagSet

type Datamodel__LogFlagpole struct {
	ByValue bool
	Details bool
}

var Datamodel__LogFlags Datamodel__LogFlagpole

func SetupDatamodel__LogFlags(fset *pflag.FlagSet, fpole *Datamodel__LogFlagpole) {
	// flags

	fset.BoolVarP(&(fpole.ByValue), "by-value", "", false, "display snapshot log by value")
	fset.BoolVarP(&(fpole.Details), "details", "", false, "print more when displaying the log")
}

func init() {
	Datamodel__LogFlagSet = pflag.NewFlagSet("Datamodel__Log", pflag.ContinueOnError)

	SetupDatamodel__LogFlags(Datamodel__LogFlagSet, &Datamodel__LogFlags)

}
