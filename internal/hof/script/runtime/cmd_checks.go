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

import (
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/opentofu/opentofu/internal/hof/lib/gotils/intern/os/execpath"
	"github.com/opentofu/opentofu/internal/hof/lib/gotils/par"
	"github.com/opentofu/opentofu/internal/hof/lib/gotils/testenv"
)

// status checks the exit or status code from the last exec or http call
func (ts *Script) CmdStatus(neg int, args []string) {
	if len(args) != 1 {
		ts.Fatalf("usage: status <int>")
	}

	// Don't care
	if neg < 0 {
		return
	}

	// Check arg
	code, err := strconv.Atoi(args[0])
	if err != nil {
		ts.Fatalf("error: %v\nusage: stdin <int>", err)
	}

	// wanted different but got samd
	if neg > 0 && ts.status == code {
		ts.Fatalf("unexpected status match: %d", code)
	}

	if neg == 0 && ts.status != code {
		ts.Fatalf("unexpected status mismatch:  wated: %d  got %d", code, ts.status)
	}

}

var execCache par.Cache[string, any]

// condition reports whether the given condition is satisfied.
func (ts *Script) condition(cond string) (bool, error) {
	switch cond {
	case "short":
		return testing.Short(), nil
	case "net":
		return testenv.HasExternalNetwork(), nil
	case "link":
		return testenv.HasLink(), nil
	case "symlink":
		return testenv.HasSymlink(), nil
	case goruntime.GOOS, goruntime.GOARCH:
		return true, nil
	default:
		if strings.HasPrefix(cond, "exec:") {
			prog := cond[len("exec:"):]
			ok := execCache.Do(prog, func() interface{} {
				_, err := execpath.Look(prog, ts.Getenv)
				return err == nil
			}).(bool)
			return ok, nil
		}
		if ts.params.Condition != nil {
			return ts.params.Condition(cond)
		}
		ts.Fatalf("unknown condition %q", cond)
		panic("unreachable")
	}
}
