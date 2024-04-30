/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package configdir_test

import (
	"os"
	"testing"

	"github.com/opentofu/opentofu/internal/hof/lib/yagu/configdir"
)

// TestCase provides common inputs and outputs for the functions being tested.
type TestCase struct {
	Env     string   // Value to put into the relevant environment variable, when Refresh=true
	Refresh bool     // Whether to call the Refresh() function before running
	Paths   []string // Path suffixes to give to the path function
	Values  []string // What we expect the return value(s) to contain
}

// On init, call the reset() function provided by the various OS-specific tests
// to reset environment variables to a known deterministic state.
func init() {
	reset()
}

// Common logic for the local paths, which return single values.
//
// Parameters:
//      t (*testing.T)
//      pathType (string): "config" or "cache", controls which environment
//          variable to play with and which path function to call.
//      defaultPrefix (string): the default path prefix for the kind of path
//          being tested, e.g. "/home/user/.config" for config paths.
//      customPrefix (string): when a custom value is set for the environment
//          variable, this is that path prefix instead of the default.
func testLocalCommon(t *testing.T, pathType, defaultPrefix, customPrefix string) {
	reset()

	// Cases to test.
	var tests = []TestCase{
		TestCase{
			Paths:  []string{},
			Values: []string{defaultPrefix},
		},
		TestCase{
			Paths:  []string{"vendor-name"},
			Values: []string{defaultPrefix + "/vendor-name"},
		},
		TestCase{
			Paths:  []string{"vendor-name", "app-name"},
			Values: []string{defaultPrefix + "/vendor-name/app-name"},
		},

		// With custom XDG paths...
		TestCase{
			Env:     customPrefix,
			Refresh: true,
			Paths:   []string{},
			Values:  []string{customPrefix},
		},
		TestCase{
			Paths:  []string{"vendor-name"},
			Values: []string{customPrefix + "/vendor-name"},
		},
		TestCase{
			Paths:  []string{"vendor-name", "app-name"},
			Values: []string{customPrefix + "/vendor-name/app-name"},
		},
	}

	for _, test := range tests {
		if test.Refresh {
			if pathType == "config" {
				os.Setenv("XDG_CONFIG_HOME", test.Env)
			} else {
				os.Setenv("XDG_CACHE_HOME", test.Env)
			}
			configdir.Refresh()
		}

		var result string
		if pathType == "config" {
			result = configdir.LocalConfig(test.Paths...)
		} else {
			result = configdir.LocalCache(test.Paths...)
		}

		if result != test.Values[0] {
			t.Errorf("Got wrong path result. Expected %s, got %s\n",
				test.Values[0],
				result,
			)
		}
	}
}
