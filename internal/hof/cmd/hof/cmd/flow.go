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
	"fmt"
	"os"

	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/cmd/flow"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/ga"

	"github.com/opentofu/opentofu/internal/hof/flow/cmd"
)

var flowLong = `run workflows and tasks powered by CUE`

func init() {

	flags.SetupFlowPflags(FlowCmd.PersistentFlags(), &(flags.FlowPflags))

}

func FlowRun(args []string) (err error) {

	// you can safely comment this print out
	// fmt.Println("not implemented")

	err = cmd.Run(args, flags.RootPflags, flags.FlowPflags)

	return err
}

var FlowCmd = &cobra.Command{

	Use: "flow [cue files...] [@flow/name...] [+key=value]",

	Aliases: []string{
		"f",
	},

	Short: "run workflows and tasks powered by CUE",

	Long: flowLong,

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		glob := toComplete + "*"
		matches, _ := filepath.Glob(glob)
		return matches, cobra.ShellCompDirectiveDefault
	},

	Run: func(cmd *cobra.Command, args []string) {

		ga.SendCommandPath(cmd.CommandPath())

		var err error

		// Argument Parsing

		err = FlowRun(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	extra := func(cmd *cobra.Command) bool {

		return false
	}

	ohelp := FlowCmd.HelpFunc()
	ousage := FlowCmd.UsageFunc()

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
	FlowCmd.SetHelpFunc(thelp)
	FlowCmd.SetUsageFunc(tusage)

	FlowCmd.AddCommand(cmdflow.ListCmd)

}
