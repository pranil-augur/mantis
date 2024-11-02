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

package flags

import (
	"github.com/spf13/pflag"
)

var _ *pflag.FlagSet

var FeedbackFlagSet *pflag.FlagSet

type FeedbackPflagpole struct {
	Issue  bool
	Labels string
}

func SetupFeedbackPflags(fset *pflag.FlagSet, fpole *FeedbackPflagpole) {
	// pflags

	fset.BoolVarP(&(fpole.Issue), "issue", "I", false, "create an issue (discussion is default)")
	fset.StringVarP(&(fpole.Labels), "labels", "L", "feedback", "labels,comma,separated")
}

var FeedbackPflags FeedbackPflagpole

func init() {
	FeedbackFlagSet = pflag.NewFlagSet("Feedback", pflag.ContinueOnError)

	SetupFeedbackPflags(FeedbackFlagSet, &FeedbackPflags)

}
