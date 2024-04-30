/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package cuetils

import (
	"bufio"
	"fmt"
	"os"

	"cuelang.org/go/cue"
)

// TODO, improve merge strategy
func AttrToMap(A cue.Attribute) (m map[string]string) {
	m = make(map[string]string)
	for i := 0; i < A.NumArgs(); i++ {
		key, val := A.Arg(i)
		m[key] = val
	}
	return m
}

func PrintAttr(attr cue.Attribute, val cue.Value) error {
	bufStdout := bufio.NewWriter(os.Stdout)
	defer bufStdout.Flush()

	// maybe print
	if attr.Err() == nil {
		for i := 0; i < attr.NumArgs(); i++ {
			a, _ := attr.String(i)
			v := val.LookupPath(cue.ParsePath(a))
			s, err := FormatCue(v)
			if err != nil {
				fmt.Fprintln(bufStdout, "Fmt error: %s", err)
			}
			fmt.Fprintf(bufStdout, "%s: %v\n", a, s)
		}
	}

	return nil
}
