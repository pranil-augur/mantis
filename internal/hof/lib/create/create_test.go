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

package create

import (
	"testing"
)

func TestLooksLikeRepo(t *testing.T) {
	type tcase struct {
		str string
	  exp bool
	}
	cases := []tcase{
		{ exp: false, str: "./" },
		{ exp: false, str: "../" },
		{ exp: false, str: "../../.." },
		{ exp: false, str: "../../../foo" },
		{ exp: false, str: "foo" },
		{ exp: false, str: "foo.cue" },
		{ exp: false, str: "foo/bar" },
		{ exp: false, str: "foo/bar.cue" },
		{ exp: false, str: "foo/bar/baz.cue" },
		{ exp: true,  str: "github.com/opentofu/opentofu/internal/hofmod-cli" },
		{ exp: true,  str: "github.com/opentofu/opentofu/internal/hofmod-cli/creator" },
		{ exp: true,  str: "hof.io/hofmod-cli" },
	}

	for _, C := range cases {
		if r := looksLikeRepo(C.str); r != C.exp {
			t.Fatalf("for %q, expected %v but got %v", C.str, C.exp, r)
		}
	}
}
