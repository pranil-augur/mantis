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

package datamodel_test

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

func TestDatamodel(t *testing.T) {
	d := ".workdir"
	os.Remove(d)
	yagu.Mkdir(d)
	runtime.Run(t, runtime.Params{
		Dir:         "./testdata",
		Glob:        "*.txt",
		WorkdirRoot: d,
		Setup:       envSetup,
	})
}

//func TestBugcases(t *testing.T) {
	//yagu.Mkdir(".workdir/bugcases")
	//runtime.Run(t, runtime.Params{
		//Dir:         "./bugcases",
		//Glob:        "*.txt",
		//WorkdirRoot: ".workdir/bugcases",
		//Setup:       envSetup,
	//})
//}

