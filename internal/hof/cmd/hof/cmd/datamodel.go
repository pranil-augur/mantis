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
	"github.com/spf13/cobra"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/cmd/datamodel"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/ga"
)

var datamodelLong = `Data models are values or objects used in many hof processes and modules.
The "datamodel" command helps you manage them and track their change history.
At their core, they represent the models that make up your application.
The intention is to define a data model for your entire application once,
then use this source of truth to generate code from database to server to client.

Hof's schema for datamodels is minimal and flexible, allowing you to define the
shape based on your application. You can have multiple datamodels as well.
You can also control and where and how history should be tracked. This history
is included during code generation so that database migrations and functions
for converting between versions can be created.

# Examples Datamodels

-- config.cue --
package datamodel

import (
	"github.com/opentofu/opentofu/internal/hof/schema/dm"
	"github.com/opentofu/opentofu/internal/hof/schema/dm/fields"
)

// Track an entire oject
Config: dm.Object & {

	host: fields.String & { Default: "8080" }

	database: {
		host:   fields.String
		port:   fields.String
		dbconn: fields.String
	}
}

-- database.cue --
package datamodel

import (
	"github.com/opentofu/opentofu/internal/hof/schema/dm/sql"
	"github.com/opentofu/opentofu/internal/hof/schema/dm/fields"
)

// Traditional database model which maps onto tables & columns
Datamodel: sql.Datamodel & {
	// implied through definition, duplicated here for example clarity
	$hof: metadata: {
		id:   "datamodel-abc123"
		name: "MyDatamodel"
	}

	Models: {
		User: {
			Fields: {
				ID:        fields.UUID
				CreatedAt: fields.Datetime
				UpdatedAt: fields.Datetime
				DeletedAt: fields.Datetime

				email:    fields.Email
				username: fields.String
				password: fields.Password
				verified: fields.Bool
				active:   fields.Bool

				persona: fields.Enum & {
					Vals: ["guest", "user", "admin", "owner"]
					Default: "user"
				}
			}
		}
	}
}

# Example Usage   (dm is short for datamodel)

  $ hof dm list   (print known data models)
  NAME         TYPE       VERSION  STATUS  ID
  Config       object     -        ok      Config
  MyDatamodel  datamodel  -        ok      datamodel-abc123

  $ hof dm tree   (print the structure of the datamodels)

  $ hof dm diff   (prints a tree based diff of the datamodel)

  $ hof dm checkpoint -m "a message about this checkpoint"

  $ hof dm log    (prints the log of changes from latest to oldest)

  You can also use the -d & -e flags to subselect datamodels and nested values

# Learn more:
  - https://docs.hofstadter.io/getting-started/data-layer/
  - https://docs.hofstadter.io/data-modeling/`

func init() {

	flags.SetupDatamodelPflags(DatamodelCmd.PersistentFlags(), &(flags.DatamodelPflags))

}

var DatamodelCmd = &cobra.Command{

	Use: "datamodel",

	Aliases: []string{
		"dm",
	},

	Short: "manage, diff, and migrate your data models",

	Long: datamodelLong,
}

func init() {
	extra := func(cmd *cobra.Command) bool {

		return false
	}

	ohelp := DatamodelCmd.HelpFunc()
	ousage := DatamodelCmd.UsageFunc()

	help := func(cmd *cobra.Command, args []string) {

		ga.SendCommandPath(cmd.CommandPath() + " help")

		if extra(cmd) {
			return
		}
		ohelp(cmd, args)
	}
	usage := func(cmd *cobra.Command) error {
		if extra(cmd) {
			return nil
		}
		return ousage(cmd)
	}

	thelp := func(cmd *cobra.Command, args []string) {
		help(cmd, args)
	}
	tusage := func(cmd *cobra.Command) error {
		return usage(cmd)
	}
	DatamodelCmd.SetHelpFunc(thelp)
	DatamodelCmd.SetUsageFunc(tusage)

	DatamodelCmd.AddCommand(cmddatamodel.CheckpointCmd)
	DatamodelCmd.AddCommand(cmddatamodel.DiffCmd)
	DatamodelCmd.AddCommand(cmddatamodel.TreeCmd)
	DatamodelCmd.AddCommand(cmddatamodel.ListCmd)
	DatamodelCmd.AddCommand(cmddatamodel.LogCmd)

}
