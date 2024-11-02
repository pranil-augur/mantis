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

package verinfo

import (
	"runtime"
	"runtime/debug"
)

var (
	Version = "Local"
	Commit  = "Dirty"

	BuildDate = "Unknown"
	GoVersion = "Unknown"
	BuildOS   = "Unknown"
	BuildArch = "Unknown"

	// todo, look this up from deps
	CueVersion = "0.7.0"

	// this is a version we can fetch with hof mod
	// the value gets injected into templates in various places
	// the default here is set to something useful for dev
	// the release version is the same as the cli running it
	HofVersion = "v0.6.8"
)


func init() {
	info, _ := debug.ReadBuildInfo()
	GoVersion = info.GoVersion

	if Version == "Local" {
		BuildOS = runtime.GOOS
		BuildArch = runtime.GOARCH

		dirty := false
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				Commit = s.Value
			}
			if s.Key == "vcs.time" {
				BuildDate = s.Value
			}
			if s.Key == "vcs.modified" {
				if s.Value == "true" {
					dirty = true
				}
			}
		}
		if dirty {
			Commit += "+dirty"
		}
	}

	// released binary override
	if Version != "Local" {
		HofVersion = Version
	}
}
