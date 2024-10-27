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

import "github.com/spf13/pflag"

type QueryPflagpole struct {
	Query bool
}

var QueryFlagSet *pflag.FlagSet

func SetupQueryPflags(fset *pflag.FlagSet, fpole *QueryPflagpole) {
	// pflags
	fset.BoolVarP(&(fpole.Query), "query", "Q", false, "Enable query mode")
}

var QueryPflags QueryPflagpole

func init() {
	QueryFlagSet = pflag.NewFlagSet("Query", pflag.ContinueOnError)

	SetupQueryPflags(QueryFlagSet, &QueryPflags)
}
