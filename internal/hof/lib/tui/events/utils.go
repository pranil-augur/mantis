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

package events

import (
	"path"
)

func findMatch(mux map[string]func(Event), path string) string {
	n := -1
	pattern := ""
	for m := range mux {
		if !isPathMatch(m, path) {
			continue
		}
		if len(m) > n {
			pattern = m
			n = len(m)
		}
	}
	return pattern

}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

func isPathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		return false
	}
	n := len(pattern)
	return len(path) >= n && path[0:n] == pattern
}
