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

package utils

import (
	"path"
	"testing"
)

func TestSplitMod(t *testing.T) {
	tests := map[string]struct {
		mod      string
		expected string
	}{
		"simple":             {mod: "github.com/owner/repo", expected: "github.com/owner/repo"},
		"submodule":          {mod: "github.com/owner/repo/submodule", expected: "github.com/owner/repo"},
		"complex":            {mod: "gitlab.com/owner/repo.git/submodule", expected: "gitlab.com/owner/repo"},
		"subgroup":           {mod: "gitlab.com/owner/subgroup/repo.git", expected: "gitlab.com/owner/subgroup/repo"},
		"subgroup+submodule": {mod: "gitlab.com/owner/subgroup/repo.git/submodule", expected: "gitlab.com/owner/subgroup/repo"},
		"small":              {mod: "cuelang.org/go", expected: "cuelang.org/go"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rm, o, rp := parseModURL(tc.mod)
			got := path.Join(rm, o, rp)
			if got != tc.expected {
				t.Fatalf("expected: %v, got: %s", tc.expected, got)
			}
		})
	}
}
