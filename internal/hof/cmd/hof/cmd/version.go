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

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/ga"
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/verinfo"
	"github.com/opentofu/opentofu/internal/hof/lib/container"
)

const versionMessage = `hof - the high code framework

Version:     %s
Commit:      %s

BuildDate:   %s
GoVersion:   %s
CueVersion:  %s
OS_Arch:     %s_%s
ConfigDir:   %s
CacheDir:    %s
Containers:  %s

Author:      Hofstadter, Inc
License:     Apache-2.0
Homepage:    https://hofstadter.io
GitHub:      https://github.com/opentofu/opentofu/internal/hof
`

var VersionLong = `Print the build version for hof`

var VersionCmd = &cobra.Command{

	Use: "version",

	Aliases: []string{
		"ver",
	},

	Short: "print the version",

	Long: VersionLong,

	Run: func(cmd *cobra.Command, args []string) {

		configDir, _ := os.UserConfigDir()
		cacheDir, _ := os.UserCacheDir()

		err := container.InitClient()
		if err != nil {
			fmt.Println(err)
		}

		rt, err := container.GetVersion()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf(
			versionMessage,
			verinfo.Version,
			verinfo.Commit,
			verinfo.BuildDate,
			verinfo.GoVersion,
			verinfo.CueVersion,
			verinfo.BuildOS,
			verinfo.BuildArch,
			filepath.Join(configDir,"hof"),
			filepath.Join(cacheDir,"hof"),
			rt,
		)
	},
}

func init() {
	help := VersionCmd.HelpFunc()
	usage := VersionCmd.UsageFunc()

	thelp := func(cmd *cobra.Command, args []string) {
		if VersionCmd.Name() == cmd.Name() {
			ga.SendCommandPath("version help")
		}
		help(cmd, args)
	}
	tusage := func(cmd *cobra.Command) error {
		if VersionCmd.Name() == cmd.Name() {
			ga.SendCommandPath("version usage")
		}
		return usage(cmd)
	}
	VersionCmd.SetHelpFunc(thelp)
	VersionCmd.SetUsageFunc(tusage)

}
