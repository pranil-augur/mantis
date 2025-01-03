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

package mmap

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"internal/syscall/windows"
)

func mmapFile(f *os.File) (Data, error) {
	st, err := f.Stat()
	if err != nil {
		return Data{}, err
	}
	size := st.Size()
	if size == 0 {
		return Data{f, nil}, nil
	}
	h, err := syscall.CreateFileMapping(syscall.Handle(f.Fd()), nil, syscall.PAGE_READONLY, 0, 0, nil)
	if err != nil {
		return Data{}, fmt.Errorf("CreateFileMapping %s: %w", f.Name(), err)
	}

	addr, err := syscall.MapViewOfFile(h, syscall.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		return Data{}, fmt.Errorf("MapViewOfFile %s: %w", f.Name(), err)
	}
	var info windows.MemoryBasicInformation
	err = windows.VirtualQuery(addr, &info, unsafe.Sizeof(info))
	if err != nil {
		return Data{}, fmt.Errorf("VirtualQuery %s: %w", f.Name(), err)
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(addr)), int(info.RegionSize))
	return Data{f, data}, nil
}
