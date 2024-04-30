/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package main

/*
Where's your docs doc?!
*/
type Context struct {
	Obj  interface{} `json:"obj" xml:"obj" yaml:"obj" form:"obj" query:"obj" `
	Path []string    `json:"path" xml:"path" yaml:"path" form:"path" query:"path" `
	Data interface{} `json:"data" xml:"data" yaml:"data" form:"data" query:"data" `
}

func NewContext() *Context {
	return &Context{
		Path: []string{},
	}
}
