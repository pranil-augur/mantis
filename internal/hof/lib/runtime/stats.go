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
	"bytes"
	"fmt"
	"time"
)

type RuntimeStats map[string]time.Duration

func (S RuntimeStats) Add(name string, dur time.Duration) {
	S[name] = dur
}
func (S RuntimeStats) String() string {
	var b bytes.Buffer

	order := []string{
		"cue/load",
		"data/load",
		"gen/load",
		"gen/run",
		"enrich/data",
		"enrich/gen",
		// "enrich/flow",
	}

	for _, o := range order {
		d, _ := S[o]
		fmt.Fprintf(&b, "%-16s%v\n", o, d.Round(time.Millisecond))
	}
	return b.String()
}

