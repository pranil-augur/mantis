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

package cmdmod

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/opentofu/opentofu/internal/hof/lib/mod"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/ga"
)

var linkLong = `symlink dependencies to cue.mod/pkg`

func LinkRun(args []string) (err error) {

	err = mod.Link(flags.RootPflags)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return err
}

var LinkCmd = &cobra.Command{

	Use: "link",

	Short: "symlink dependencies to cue.mod/pkg",

	Long: linkLong,

	Run: func(cmd *cobra.Command, args []string) {

		ga.SendCommandPath(cmd.CommandPath())

		var err error

		// Argument Parsing

		err = LinkRun(args)
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

	ohelp := LinkCmd.HelpFunc()
	ousage := LinkCmd.UsageFunc()

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
	LinkCmd.SetHelpFunc(thelp)
	LinkCmd.SetUsageFunc(tusage)

}
