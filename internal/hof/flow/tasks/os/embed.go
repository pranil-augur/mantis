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

package os

import (
	_ "embed"
	"fmt"

	"cuelang.org/go/cue"
)

//go:embed schema.cue
var task_schema string

var task_chan cue.Value
var task_send cue.Value
var task_recv cue.Value
var task_watch cue.Value

func init_schemas(ctx *cue.Context) {
	if task_chan.Exists() {
		return
	}

	val := ctx.CompileString(task_schema, cue.Filename("@embed:flow/tasks/csp/schema.cue"))
	if val.Err() != nil {
		fmt.Println(val.Err())
		panic("should not have a schema error")
	}

	task_chan = val.LookupPath(cue.ParsePath("Chan"))
	if !task_chan.Exists() {
		panic("missing flow/tasks/csp.Chan schema")
	}
	task_send = val.LookupPath(cue.ParsePath("Send"))
	if !task_send.Exists() {
		panic("missing flow/tasks/csp.Send schema")
	}
	task_recv = val.LookupPath(cue.ParsePath("Recv"))
	if !task_recv.Exists() {
		panic("missing flow/tasks/csp.Recv schema")
	}

	task_watch = val.LookupPath(cue.ParsePath("Watch"))
	if !task_watch.Exists() {
		panic("missing flow/tasks/fs.Watch schema")
	}
}
