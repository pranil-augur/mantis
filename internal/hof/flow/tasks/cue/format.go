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

package cue

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type CueFormat struct {
	Package        string
	Raw            bool
	Final          bool
	Concrete       bool
	Definitions    bool
	Optional       bool
	Hidden         bool
	Attributes     bool
	Docs           bool
	InlineImports  bool
	ErrorsAsValues bool
}

func NewCueFormat(val cue.Value) (hofcontext.Runner, error) {
	return &CueFormat{}, nil
}

func (T *CueFormat) Run(ctx *hofcontext.Context) (interface{}, error) {

	v := ctx.Value
	var val cue.Value

	ferr := func() error {
		ctx.CUELock.Lock()
		defer func() {
			ctx.CUELock.Unlock()
		}()

		err := v.Decode(T)
		if err != nil {
			return err
		}

		val = v.LookupPath(cue.ParsePath("value"))
		if !val.Exists() {
			return fmt.Errorf("in task %s: missing field 'value'", v.Path())
		}
		if val.Err() != nil {
			return val.Err()
		}

		return nil
	}()
	if ferr != nil {
		return nil, ferr
	}

	opts := []cue.Option{
		cue.Concrete(T.Concrete),
		cue.Definitions(T.Definitions),
		cue.Optional(T.Optional),
		cue.Hidden(T.Hidden),
		cue.Attributes(T.Attributes),
		cue.Docs(T.Docs),
		// cue.InlineImports(T.InlineImports),
		cue.ErrorsAsValues(T.ErrorsAsValues),
	}
	if T.Final {
		opts = append(opts, cue.Final())
	}
	if T.Raw {
		opts = append(opts, cue.Raw())
	}

	syn := val.Syntax(opts...)

	if T.Package != "" {
		pkgDecl := &ast.Package{
			Name: ast.NewIdent(T.Package),
		}
		decls := []ast.Decl{pkgDecl}
		// this could cause an issue?
		switch t := syn.(type) {
		case *ast.File:
			t.Decls = append(decls, t.Decls...)

		case *ast.StructLit:
			decls = append(decls, t.Elts...)
			f := &ast.File{
				Decls: decls,
			}
			syn = f
		case *ast.ListLit:
			decls = append(decls, t)
			f := &ast.File{
				Decls: decls,
			}
			syn = f
		}
	}

	bs, err := format.Node(syn)
	if err != nil {
		return nil, err
	}

	ctx.CUELock.Lock()
	defer ctx.CUELock.Unlock()
	res := v.FillPath(cue.ParsePath("out"), string(bs))

	return res, nil
}
