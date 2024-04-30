/*
 * Augur AI Proprietary
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

var RunFlagSet *pflag.FlagSet

type RunFlagpole struct {
	Mode        string
	Workdir     string
	KeepTestdir bool
}

var RunFlags RunFlagpole

func SetupRunFlags(fset *pflag.FlagSet, fpole *RunFlagpole) {
	// flags

	fset.StringVarP(&(fpole.Mode), "mode", "m", "run", "set the script execution mode")
	fset.StringVarP(&(fpole.Workdir), "workdir", "w", "", "working directory")
	fset.BoolVarP(&(fpole.KeepTestdir), "keep-testdir", "", false, "keep the workdir after test mode run")
}

func init() {
	RunFlagSet = pflag.NewFlagSet("Run", pflag.ContinueOnError)

	SetupRunFlags(RunFlagSet, &RunFlags)

}
