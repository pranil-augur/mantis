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

package mod_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
	"github.com/opentofu/opentofu/internal/hof/script/runtime"
)

func envSetup(env *runtime.Env) error {

	vars := []string{
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"BITBUCKET_USERNAME",
		"BITBUCKET_PASSWORD",

		"HOF_FMT_VERSION",
		"DOCKER_HOST",
		"CONTAINERD_ADDRESS",
		"CONTAINERD_NAMESPACE",
	}
	for _, v := range vars {
		if val := os.Getenv(v); val != "" {
			env.Vars = append(env.Vars, fmt.Sprintf("%s=%s", v, val))
		}
	}
	env.Vars = append(env.Vars, "HOF_TELEMETRY_DISABLED=1")
	return nil
}

func setupWorkdir(dir string) {
	os.RemoveAll(dir)
	yagu.Mkdir(dir)
}

func TestModTests(t *testing.T) {
	d := ".workdir/tests"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata",
		Glob:        "*.txt",
		WorkdirRoot: d,
	})
}

/*
func TestModBugs(t *testing.T) {
	d := ".workdir/bugs"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata/bugs",
		Glob:        "*.txt",
		WorkdirRoot: d,
	})
}
*/

func TestModAuthdApikeysTests(t *testing.T) {
	d := ".workdir/authd/apikeys"
	setupWorkdir(d)
	runtime.Run(t, runtime.Params{
		Setup:       envSetup,
		Dir:         "testdata/authd/apikeys",
		Glob:        "*.txt",
		WorkdirRoot: d,
	})
}

// we don't support sshkey auth at this point
// libraries are giving issue

//func TestModAuthdSshconfigTests(t *testing.T) {
//  d := ".workdir/authd/sshconfig"
//  setupWorkdir(d)
//  runtime.Run(t, runtime.Params{
//    Setup:       envSetup,
//    Dir:         "testdata/authd/sshconfig",
//    Glob:        "*.txt",
//    WorkdirRoot: d,
//  })
//}

//func TestModAuthdSshkeyTests(t *testing.T) {
//  d := ".workdir/authd/sshkey"
//  setupWorkdir(d)
//  runtime.Run(t, runtime.Params{
//    Setup: envSetup,
//    Dir: "testdata/authd/sshkey",
//    Glob: "*.txt",
//    WorkdirRoot: d,
//  })
//}
