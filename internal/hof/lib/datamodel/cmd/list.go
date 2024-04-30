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
	// "fmt"

	"github.com/codemodus/kace"
	"github.com/olekukonko/tablewriter"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/datamodel"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func list(R *runtime.Runtime, dflags flags.DatamodelPflagpole) error {
	return printAsTable(
		[]string{"Name", "Type", "Version", "Status", "ID"},
		func(table *tablewriter.Table) ([][]string, error) {
			var rows = make([][]string, 0, len(R.Datamodels))
			// fill with data
			for _, dm := range R.Datamodels {
				id := dm.Hof.Metadata.ID
				if id == "" {
					id = kace.Snake(dm.Hof.Metadata.Name) + " (auto)"
				}

				name := dm.Hof.Metadata.Name
				typ  := datamodel.DatamodelType(dm)
				ver := dm.Hof.Datamodel.Version
				if ver == "" {
					ver = "-"
				}
				status := dm.Status()
				if status == "" {
					status = "-"
				}

				row := []string{name, typ, ver, status, id}
				rows = append(rows, row)
			}
			return rows, nil
		},
	)
}
