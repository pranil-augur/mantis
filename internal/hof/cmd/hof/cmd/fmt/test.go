/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package cmdfmt

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/ga"

	hfmt "github.com/opentofu/opentofu/internal/hof/lib/fmt"
)

var testLong = `test that formatter(s) are working`

func TestRun(formatter string) (err error) {

	// you can safely comment this print out
	// fmt.Println("not implemented")

	err = hfmt.Test(formatter)

	return err
}

var TestCmd = &cobra.Command{

	Use: "test",

	Short: "test that formatter(s) are working",

	Long: testLong,

	Run: func(cmd *cobra.Command, args []string) {

		ga.SendCommandPath(cmd.CommandPath())

		var err error

		// Argument Parsing

		var formatter string

		if 0 < len(args) {

			formatter = args[0]

		}

		err = TestRun(formatter)
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

	ohelp := TestCmd.HelpFunc()
	ousage := TestCmd.UsageFunc()

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
	TestCmd.SetHelpFunc(thelp)
	TestCmd.SetUsageFunc(tusage)

}
