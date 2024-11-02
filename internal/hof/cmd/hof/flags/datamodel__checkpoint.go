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

var Datamodel__CheckpointFlagSet *pflag.FlagSet

type Datamodel__CheckpointFlagpole struct {
	Message string
}

var Datamodel__CheckpointFlags Datamodel__CheckpointFlagpole

func SetupDatamodel__CheckpointFlags(fset *pflag.FlagSet, fpole *Datamodel__CheckpointFlagpole) {
	// flags

	fset.StringVarP(&(fpole.Message), "message", "m", "", "message describing the checkpoint")
}

func init() {
	Datamodel__CheckpointFlagSet = pflag.NewFlagSet("Datamodel__Checkpoint", pflag.ContinueOnError)

	SetupDatamodel__CheckpointFlags(Datamodel__CheckpointFlagSet, &Datamodel__CheckpointFlags)

}
