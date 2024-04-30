/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package runtime

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattn/go-zglob"

	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

func keepFilter(hn *hof.Node[any], patterns []string) bool {
	// filter by name
	if len(patterns) > 0 {
		for _, d := range patterns {

			// three match variations
			// 1. regexp when /.../
			// 2. glob if any *
			// 3. string prefix
			if strings.HasPrefix(d,"/") && strings.HasSuffix(d,"/") {
				// regexp
				match, err := regexp.MatchString(d, hn.Hof.Metadata.Name)
				if err != nil {
					fmt.Println("error:", err)
					return false
				}
				if match {
					return true
				}
			} else if strings.Contains(d,"*") {
				// glob
				match, err := zglob.Match(d, hn.Hof.Metadata.Name)
				if err != nil {
					fmt.Println("error:", err)
					return false
				}
				if match {
					return true
				}
			} else {
				// prefix
				if strings.HasPrefix(hn.Hof.Metadata.Name, d) {
					return true	
				}
			}
		}
		return false
	}

	// filter by time

	// filter by version?

	// default to true, should include everything when no checks are needed
	return true
}

