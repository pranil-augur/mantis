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
	"strings"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
)

func List(args []string, rflags flags.RootPflagpole, gflags flags.GenFlagpole) error {
	R, err := prepRuntime(args, rflags, gflags)
	if err != nil {
		return err
	}

	// TODO...
	// 1. use table printer
	// 2. move this command up, large blocks of this ought
	gens := make([]string, 0, len(R.Generators))
	for _, G := range R.Generators {
		gens = append(gens, G.Hof.Metadata.Name)
	}
	if len(gens) == 0 {
		return fmt.Errorf("no generators found")
	}
	fmt.Printf("Available Generators\n  ")
	fmt.Println(strings.Join(gens, "\n  "))
	
	// print gens
	return nil
}
