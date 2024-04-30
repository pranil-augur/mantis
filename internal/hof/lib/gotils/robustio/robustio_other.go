/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows && !darwin

package robustio

import (
	"os"
)

func rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func readFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func removeAll(path string) error {
	return os.RemoveAll(path)
}

func isEphemeralError(err error) bool {
	return false
}
