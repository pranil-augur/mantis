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

func List(args []string, rflags flags.RootPflagpole) error {
	R, err := prepRuntime(args, rflags)
	if err != nil {
		return err
	}

	// TODO...
	// 1. use table printer
	// 2. move this command up, large blocks of this ought
	chats := make([]string, 0, len(R.Chats))
	for _, C := range R.Chats {
		chats = append(chats, C.Hof.Metadata.Name)
	}
	if len(chats) == 0 {
		return fmt.Errorf("no chats found")
	}
	fmt.Printf("Available Chats\n  ")
	fmt.Println(strings.Join(chats, "\n  "))
	
	// print gens
	return nil
}
