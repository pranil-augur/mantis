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

var Gen__InfoFlagSet *pflag.FlagSet

type Gen__InfoFlagpole struct {
	Expression []string
}

var Gen__InfoFlags Gen__InfoFlagpole

func SetupGen__InfoFlags(fset *pflag.FlagSet, fpole *Gen__InfoFlagpole) {
	// flags

	fset.StringArrayVarP(&(fpole.Expression), "expr", "e", nil, "CUE paths to select outputs, depending on the command")
}

func init() {
	Gen__InfoFlagSet = pflag.NewFlagSet("Gen__Info", pflag.ContinueOnError)

	SetupGen__InfoFlags(Gen__InfoFlagSet, &Gen__InfoFlags)

}
