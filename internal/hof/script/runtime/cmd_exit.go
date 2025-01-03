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

// skip marks the test skipped.
func (ts *Script) CmdSkip(neg int, args []string) {
	if neg != 0 {
		ts.Fatalf("unsupported: !? skip")
	}

	if len(args) > 1 {
		ts.Fatalf("usage: skip [msg]")
	}

	// Before we mark the test as skipped, shut down any background processes and
	// make sure they have returned the correct status.
	for _, bg := range ts.background {
		interruptProcess(bg.cmd.Process)
	}
	ts.CmdWait(0, nil)

	if len(args) == 1 {
		ts.t.Skip(args[0])
	}
	ts.t.Skip()
	ts.stopped = true
}

// should we stop background in CmdStop? probably

// stop stops execution of the test (marking it passed).
func (ts *Script) CmdStop(neg int, args []string) {
	if neg != 0 {
		ts.Fatalf("unsupported: !? stop")
	}
	if len(args) > 1 {
		ts.Fatalf("usage: stop [msg]")
	}
	if len(args) == 1 {
		ts.Logf("stop: %s\n", args[0])
	} else {
		ts.Logf("stop\n")
	}
	ts.stopped = true
}
