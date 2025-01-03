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

// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package is a lightly modified version of the mmap code
// in github.com/google/codesearch/index.

// The mmap package provides an abstraction for memory mapping files
// on different platforms.
package mmap

import (
	"os"
)

// Data is mmap'ed read-only data from a file.
// The backing file is never closed, so Data
// remains valid for the lifetime of the process.
type Data struct {
	f    *os.File
	Data []byte
}

// Mmap maps the given file into memory.
func Mmap(file string) (Data, error) {
	f, err := os.Open(file)
	if err != nil {
		return Data{}, err
	}
	return mmapFile(f)
}
