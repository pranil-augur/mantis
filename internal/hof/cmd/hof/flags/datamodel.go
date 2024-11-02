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

var DatamodelFlagSet *pflag.FlagSet

type DatamodelPflagpole struct {
	Datamodels []string
	Expression []string
}

func SetupDatamodelPflags(fset *pflag.FlagSet, fpole *DatamodelPflagpole) {
	// pflags

	fset.StringArrayVarP(&(fpole.Datamodels), "model", "M", nil, "specify one or more data models to operate on")
	fset.StringArrayVarP(&(fpole.Expression), "expr", "e", nil, "CUE paths to select outputs, depending on the command")
}

var DatamodelPflags DatamodelPflagpole

func init() {
	DatamodelFlagSet = pflag.NewFlagSet("Datamodel", pflag.ContinueOnError)

	SetupDatamodelPflags(DatamodelFlagSet, &DatamodelPflags)

}
