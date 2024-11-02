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

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/chat"
	"github.com/opentofu/opentofu/internal/hof/lib/cuetils"
)

func Info(name string, entrypoints []string, rflags flags.RootPflagpole) error {
	R, err := prepRuntime(entrypoints, rflags)
	if err != nil {
		return err
	}

	// TODO...
	// 1. use table printer
	// 2. move this command up, large blocks of this ought
	var c *chat.Chat	
	for _, C := range R.Chats {
		if C.Hof.Metadata.Name == name {
			c = C
		}
	}
	if c == nil {
		return fmt.Errorf("no chat %q found", name)
	}

	err = c.Value.Decode(c)
	if err != nil {
		err = cuetils.ExpandCueError(err)
		return err
	}

	fmt.Println("name:        ", c.Name)
	fmt.Println("model:       ", c.Model)
	fmt.Println("description: ", c.Description)
	
	// print gens
	return nil
}
