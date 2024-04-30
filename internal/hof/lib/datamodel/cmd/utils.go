/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package cmd

import (
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func findMaxLabelLen(R *runtime.Runtime, dflags flags.DatamodelPflagpole) int {
	max := 0
	for _, dm := range R.Datamodels {
		m := dm.FindMaxLabelLen(dflags)
		if m > max {
			max = m
		}
	}
	return max
}


