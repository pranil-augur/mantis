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

//go:build unix

package mmap

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func mmapFile(f *os.File) (Data, error) {
	st, err := f.Stat()
	if err != nil {
		return Data{}, err
	}
	size := st.Size()
	pagesize := int64(os.Getpagesize())
	if int64(int(size+(pagesize-1))) != size+(pagesize-1) {
		return Data{}, fmt.Errorf("%s: too large for mmap", f.Name())
	}
	n := int(size)
	if n == 0 {
		return Data{f, nil}, nil
	}
	mmapLength := int(((size + pagesize - 1) / pagesize) * pagesize) // round up to page size
	data, err := syscall.Mmap(int(f.Fd()), 0, mmapLength, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return Data{}, &fs.PathError{Op: "mmap", Path: f.Name(), Err: err}
	}
	return Data{f, data[:n]}, nil
}
