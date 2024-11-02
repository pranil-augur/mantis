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

package cache

import (
	"os"

	"golang.org/x/mod/sumdb/dirhash"

	"github.com/opentofu/opentofu/internal/hof/lib/repos/utils"
)

func Checksum(mod, ver string) (string, error) {
	remote, owner, repo := utils.ParseModURL(mod)
	tag := ver

	dir := ModuleOutdir(remote, owner, repo, tag)

	_, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}

	h, err := dirhash.HashDir(dir, mod, dirhash.Hash1)

	return h, err
}
