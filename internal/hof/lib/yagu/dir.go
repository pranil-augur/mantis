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

package yagu

import (
	"os"
	"path/filepath"
)

// cleanup empty dirs, walk up
func RemoveEmptyDirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	// if entries, we can return
	if len(entries) > 0 {
		return nil
	}

	// remove dir
	err = os.Remove(dir)
	if err != nil {
		return err
	}

	// recurse to parent dir
	return RemoveEmptyDirs(filepath.Dir(dir))
}
