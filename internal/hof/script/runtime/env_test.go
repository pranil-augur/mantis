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
	"testing"
)

func TestEnv(t *testing.T) {
	e := &Env{
		Vars: []string{
			"HOME=/no-home",
			"PATH=/usr/bin",
			"PATH=/usr/bin:/usr/local/bin",
			"INVALID",
		},
	}

	if got, want := e.Getenv("HOME"), "/no-home"; got != want {
		t.Errorf("e.Getenv(\"HOME\") == %q, want %q", got, want)
	}

	e.Setenv("HOME", "/home/user")
	if got, want := e.Getenv("HOME"), "/home/user"; got != want {
		t.Errorf(`e.Getenv("HOME") == %q, want %q`, got, want)
	}

	if got, want := e.Getenv("PATH"), "/usr/bin:/usr/local/bin"; got != want {
		t.Errorf(`e.Getenv("PATH") == %q, want %q`, got, want)
	}

	if got, want := e.Getenv("INVALID"), ""; got != want {
		t.Errorf(`e.Getenv("INVALID") == %q, want %q`, got, want)
	}

	for _, key := range []string{
		"",
		"=",
		"key=invalid",
	} {
		var panicValue interface{}
		func() {
			defer func() {
				panicValue = recover()
			}()
			e.Setenv(key, "")
		}()
		if panicValue == nil {
			t.Errorf("e.Setenv(%q) did not panic, want panic", key)
		}
	}
}
