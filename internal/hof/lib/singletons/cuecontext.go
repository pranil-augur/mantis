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

package singletons

import (
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

var cueContext *cue.Context
var cueContextMutex sync.Mutex

func init() {
	cueContext = cuecontext.New()
}

func CueContext() *cue.Context {
	return cueContext
}

func EmptyValue() cue.Value {
	cueContextMutex.Lock()
	defer cueContextMutex.Unlock()

	return cueContext.CompileString("{}")
}

func CompileString(src string, opts ...cue.BuildOption) cue.Value {
	cueContextMutex.Lock()
	defer cueContextMutex.Unlock()

	return cueContext.CompileString(src, opts...)
}

func CompileBytes(src []byte, opts ...cue.BuildOption) cue.Value {
	cueContextMutex.Lock()
	defer cueContextMutex.Unlock()

	return cueContext.CompileBytes(src, opts...)
}
