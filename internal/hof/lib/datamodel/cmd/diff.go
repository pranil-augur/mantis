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
	"os"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func diff(R *runtime.Runtime, dflags flags.DatamodelPflagpole) error {

	for _, dm := range R.Datamodels {
		if !dm.HasDiff() {
			continue
		}
		if err := dm.PrintDiff(os.Stdout, dflags); err != nil {
			return err
		}
	}

	return nil
}
