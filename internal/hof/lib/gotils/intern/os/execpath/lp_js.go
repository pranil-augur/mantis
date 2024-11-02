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

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js && wasm
// +build js,wasm

package execpath

// Look searches for an executable named file, using getenv to look up
// environment variables. If getenv is nil, os.Getenv will be used. If file
// contains a slash, it is tried directly and getenv will not be called.  The
// result may be an absolute path or a path relative to the current directory.
func Look(file string, getenv func(string) string) (string, error) {
	// Wasm can not execute processes, so act as if there are no executables at all.
	return "", &Error{file, ErrNotFound}
}
