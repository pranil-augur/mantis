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
	"fmt"
	"time"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/datamodel"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func checkpoint(R *runtime.Runtime, dflags flags.DatamodelPflagpole, cflags flags.Datamodel__CheckpointFlagpole) error {
	timestamp := time.Now().UTC().Format(datamodel.CheckpointTimeFmt)
	fmt.Printf("creating checkpoint: %s %q\n", timestamp, cflags.Message)

	for _, dm := range R.Datamodels {
		err := dm.MakeSnapshot(timestamp, dflags, cflags)
		if err != nil {
			return err
		}
	}

	return nil
}
