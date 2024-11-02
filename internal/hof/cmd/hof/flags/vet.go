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

var VetFlagSet *pflag.FlagSet

type VetFlagpole struct {
	Expression  []string
	List        bool
	Simplify    bool
	Out         string
	Outfile     string
	Concrete    bool
	Comments    bool
	Attributes  bool
	Definitions bool
	Hidden      bool
	Optional    bool
}

var VetFlags VetFlagpole

func SetupVetFlags(fset *pflag.FlagSet, fpole *VetFlagpole) {
	// flags

	fset.StringArrayVarP(&(fpole.Expression), "expression", "e", nil, "evaluate these expressions only")
	fset.BoolVarP(&(fpole.List), "list", "", false, "concatenate multiple objects into a list")
	fset.BoolVarP(&(fpole.Simplify), "simplify", "", false, "simplify CUE statements where possible")
	fset.StringVarP(&(fpole.Out), "out", "", "", "output data format, when detection does not work")
	fset.StringVarP(&(fpole.Outfile), "outfile", "o", "", "filename or - for stdout with optional file prefix")
	fset.BoolVarP(&(fpole.Concrete), "concrete", "c", false, "require the evaluation to be concrete")
	fset.BoolVarP(&(fpole.Comments), "comments", "C", false, "include comments in output")
	fset.BoolVarP(&(fpole.Attributes), "attributes", "A", false, "display field attributes")
	fset.BoolVarP(&(fpole.Definitions), "definitions", "S", true, "display defintions")
	fset.BoolVarP(&(fpole.Hidden), "hidden", "H", false, "display hidden fields")
	fset.BoolVarP(&(fpole.Optional), "optional", "O", false, "display optional fields")
}

func init() {
	VetFlagSet = pflag.NewFlagSet("Vet", pflag.ContinueOnError)

	SetupVetFlags(VetFlagSet, &VetFlags)

}
