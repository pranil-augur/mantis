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

package yagu

import (
	"reflect"
	"testing"
)

var testNetrc = `
machine incomplete
password none

machine api.github.com
  login user
  password pwd

machine incomlete.host
  login justlogin

machine test.host
login user2
password pwd2

machine oneline login user3 password pwd3

machine ignore.host macdef ignore
  login nobody
  password nothing

machine hasmacro.too macdef ignore-next-lines login user4 password pwd4
  login nobody
  password nothing

default
login anonymous
password gopher@golang.org

machine after.default
login oops
password too-late-in-file
`

func TestParseNetrc(t *testing.T) {
	lines := parseNetrc(testNetrc)
	want := map[string]NetrcMachine{
		"api.github.com": {"user", "pwd"},
		"test.host":      {"user2", "pwd2"},
		"oneline":        {"user3", "pwd3"},
		"hasmacro.too":   {"user4", "pwd4"},
	}

	if !reflect.DeepEqual(lines, want) {
		t.Errorf("parseNetrc:\nhave %q\nwant %q", lines, want)
	}
}
