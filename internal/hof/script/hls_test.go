/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package script_test

import (
	"testing"

	"github.com/opentofu/opentofu/internal/hof/script/runtime"
)

func TestScriptBrowser(t *testing.T) {
	runtime.Run(t, runtime.Params{
		Dir:  "tests/browser",
		Glob: "*.hls",
	})
}

func TestScriptCmds(t *testing.T) {
	runtime.Run(t, runtime.Params{
		Dir:  "tests/cmds",
		Glob: "*.hls",
	})
}

func TestScriptHTTP(t *testing.T) {
	runtime.Run(t, runtime.Params{
		Dir:  "tests/http",
		Glob: "*.hls",
	})
}
