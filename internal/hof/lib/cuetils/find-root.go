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

package cuetils

import (
	"os"
	"path/filepath"
)

func FindModuleAbsPath(dir string) (string, error) {
	var err error
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	found := false

	for !found && dir != "/" {
		try := filepath.Join(dir, "cue.mod")
		info, err := os.Stat(try)
		if err == nil && info.IsDir() {
			found = true
			break
		}

		next := filepath.Clean(filepath.Join(dir, ".."))
		dir = next
	}

	if !found {
		return "", nil
		// return "", fmt.Errorf("unable to find CUE module root")
	}

	return dir, nil
}
