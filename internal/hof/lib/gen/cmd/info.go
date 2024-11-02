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

package cmd

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
)

func Info(args []string, rflags flags.RootPflagpole, gflags flags.GenFlagpole, iflags flags.Gen__InfoFlagpole) error {
	R, err := prepRuntime(args, rflags, gflags)
	if err != nil {
		return err
	}

	if len(iflags.Expression) == 0 {
		fmt.Println(R.Value)
		return nil
	}

	for _, ex := range iflags.Expression {
		val := R.Value.LookupPath(cue.ParsePath(ex))
		path := val.Path()
		fmt.Printf("%s: %v\n\n", path, val)
	}

	return nil


	//for _, G := range R.Generators {
		//if len(iflags.Expression) == 0 {
			//fmt.Printf("%s: %v\n\n", G.Hof.Metadata.Name, G.CueValue)
			//continue
		//}

		//for _, ex := range iflags.Expression {
			//val := G.CueValue.LookupPath(cue.ParsePath(ex))
			//path := G.Hof.Metadata.Name + "." + ex
			//fmt.Printf("%s: %v\n\n", path, val)
		//}
	//}

	// return nil
}

