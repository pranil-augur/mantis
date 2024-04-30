/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package api

import (
	_ "embed"
	"fmt"

	"cuelang.org/go/cue"
)

//go:embed schema.cue
var task_schema string

var task_call cue.Value
var task_serve cue.Value

func init_schemas(ctx *cue.Context) {
	if task_call.Exists() {
		return
	}

	val := ctx.CompileString(task_schema, cue.Filename("@embed:flow/tasks/api/schema.cue"))
	if val.Err() != nil {
		fmt.Println(val.Err())
		panic("should not have a schema error")
	}

	task_call = val.LookupPath(cue.ParsePath("Call"))
	if !task_call.Exists() {
		panic("missing flow/tasks/api.Call schema")
	}
	task_serve = val.LookupPath(cue.ParsePath("Serve"))
	if !task_serve.Exists() {
		panic("missing flow/tasks/api.Serve schema")
	}
}
