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

package cuecmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
	"github.com/opentofu/opentofu/internal/hof/script/runtime"
)

func envSetup(env *runtime.Env) error {
	env.Vars = append(env.Vars, "HOF_TELEMETRY_DISABLED=1")

	vars := []string{
		"GITHUB_TOKEN",
		"HOF_FMT_VERSION",
		"DOCKER_HOST",
		"CONTAINERD_ADDRESS",
		"CONTAINERD_NAMESPACE",
	}

	for _,v := range vars {
		val := os.Getenv(v)
		jnd := fmt.Sprintf("%s=%s", v, val)
		env.Vars = append(env.Vars, jnd)
	}

	return nil
}

func setupWorkdir(dir string) {
	os.RemoveAll(dir)
	yagu.Mkdir(dir)
}

func TestDef(t *testing.T) {
	d := ".workdir/def"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata",
		Glob:        "def_*.txt",
		WorkdirRoot: d,
	})
}

func TestEval(t *testing.T) {
	d := ".workdir/eval"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata",
		Glob:        "eval_*.txt",
		WorkdirRoot: d,
	})
}

func TestExport(t *testing.T) {
	d := ".workdir/export"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata",
		Glob:        "export_*.txt",
		WorkdirRoot: d,
	})
}

func TestVet(t *testing.T) {
	d := ".workdir/vet"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata",
		Glob:        "vet_*.txt",
		WorkdirRoot: d,
	})
}

