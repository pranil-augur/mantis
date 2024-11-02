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

var FmtFlagSet *pflag.FlagSet

type FmtFlagpole struct {
	Data bool
}

var FmtFlags FmtFlagpole

func SetupFmtFlags(fset *pflag.FlagSet, fpole *FmtFlagpole) {
	// flags

	fset.BoolVarP(&(fpole.Data), "fmt-data", "", true, "include cue,yaml,json,toml,xml files, set to false to disable")
}

func init() {
	FmtFlagSet = pflag.NewFlagSet("Fmt", pflag.ContinueOnError)

	SetupFmtFlags(FmtFlagSet, &FmtFlags)

}
