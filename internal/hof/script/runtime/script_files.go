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
	"io/ioutil"

	"github.com/opentofu/opentofu/internal/hof/lib/gotils/txtar"
)

// unquote unquotes files.
func (ts *Script) CmdUnquote(neg int, args []string) {
	if neg != 0 {
		ts.Fatalf("unsupported: !? unquote")
	}
	for _, arg := range args {
		file := ts.MkAbs(arg)
		data, err := ioutil.ReadFile(file)
		ts.Check(err)
		data, err = txtar.Unquote(data)
		ts.Check(err)
		err = ioutil.WriteFile(file, data, 0666)
		ts.Check(err)
	}
}
