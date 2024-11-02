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

//go:build !windows && !darwin
// +build !windows,!darwin

package configdir

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	systemConfig []string
	localConfig  string
	localCache   string
)

func findPaths() {
	// System-wide configuration.
	if os.Getenv("XDG_CONFIG_DIRS") != "" {
		systemConfig = strings.Split(os.Getenv("XDG_CONFIG_DIRS"), ":")
	} else {
		systemConfig = []string{"/etc/xdg"}
	}

	// Local user configuration.
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		localConfig = os.Getenv("XDG_CONFIG_HOME")
	} else {
		localConfig = filepath.Join(os.Getenv("HOME"), ".config")
	}

	// Local user cache.
	if os.Getenv("XDG_CACHE_HOME") != "" {
		localCache = os.Getenv("XDG_CACHE_HOME")
	} else {
		localCache = filepath.Join(os.Getenv("HOME"), ".cache")
	}
}
