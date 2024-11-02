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

package script

import (
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5/osfs"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/script/_ast"
)

func Hack(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("please supply a single filepath")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	fs := osfs.New(cwd)

	llvls := []string{"error", "warn", "info", "debug"}
	llvl := llvls[flags.RootPflags.Verbosity]

	config := &ast.Config{
		LogLevel: llvl,
		FS:       fs,
	}
	parser := ast.NewParser(config)

	S, err := parser.ParseScript(args[0])
	if err != nil {
		fmt.Println("ERROR:", err)
		return err
	}

	fmt.Println("done hacking ", S.Path)

	return nil
}
