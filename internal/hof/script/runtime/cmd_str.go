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

package runtime

// regexp checks that file content matches a regexp.
// it accepts Go regexp syntax.
func (ts *Script) CmdRegexp(neg int, args []string) {
	scriptMatch(ts, neg, args, "", "regexp")
}

// regexp checks that file content matches a regexp.
// it accepts Go regexp syntax and returns the matches
func (ts *Script) CmdGrep(neg int, args []string) {
	scriptMatch(ts, neg, args, "", "grep")
}

// sed finds and replaces in text content
// it accepts Go regexp syntax and returns the replaced content
func (ts *Script) CmdSed(neg int, args []string) {
	scriptMatch(ts, neg, args, "", "sed")
}
