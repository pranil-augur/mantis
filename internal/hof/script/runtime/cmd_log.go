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

func (ts *Script) CmdLog(neg int, args []string) {
	if neg != 0 {
		ts.Fatalf("unsupported: !? log")
	}

	if len(args) < 2 {
		ts.Fatalf("usage: log <lvl|set> ...")
	}

	cmd, msg, rest := args[0], args[1], args[2:]
	// convert rest for logging
	irest := make([]interface{}, len(rest), len(rest))
	for i := range rest {
		irest[i] = rest[i]
	}

	switch cmd {
	case "debug":
		ts.Logger.Debugw(msg, irest...)
	case "debugf":
		ts.Logger.Debugf(msg, irest...)

	case "info":
		ts.Logger.Infow(msg, irest...)
	case "infof":
		ts.Logger.Infof(msg, irest...)

	case "warn":
		ts.Logger.Warnw(msg, irest...)
	case "warnf":
		ts.Logger.Warnf(msg, irest...)

	case "error":
		ts.Logger.Errorw(msg, irest...)
	case "errorf":
		ts.Logger.Errorf(msg, irest...)

	case "fatal":
		ts.Logger.Fatalw(msg, irest...)
	case "fatalf":
		ts.Logger.Fatalf(msg, irest...)

	default:
		ts.Fatalf("unknown arg %q\nusage: log <lvl|set> ...", cmd)
	}

}
