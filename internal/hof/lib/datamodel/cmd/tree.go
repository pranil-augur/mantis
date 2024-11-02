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

package cmd

import (
	"os"

	/*
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/format"
	*/

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func tree(R *runtime.Runtime, dflags flags.DatamodelPflagpole) error {

	// find max label width after indentation for column alignment
	max := findMaxLabelLen(R, dflags)

	for _, dm := range R.Datamodels {
		if err := dm.PrintTree(os.Stdout, max, dflags); err != nil {
			return err
		}

		/*

		name := dm.Hof.Label
		p := cue.ParsePath(name)

		ctx := dm.Value.Context()
		val := ctx.CompileString("_")

		val = val.FillPath(p, dm.Value)

		// add lacunas

		node := val.Syntax(
			cue.Final(),
			cue.Docs(true),
			cue.Attributes(true),
			cue.Definitions(true),
			cue.Optional(true),
			cue.Hidden(true),
			cue.Concrete(true),
			cue.ResolveReferences(true),
		)

		file, err := astutil.ToFile(node.(*ast.StructLit))
		if err != nil {
			return err
		}

		pkg := &ast.Package{
			Name: ast.NewIdent("info"),
		}
		file.Decls = append([]ast.Decl{pkg}, file.Decls...)

		// fmt.Printf("%#+v\n", file)

		bytes, err := format.Node(
			file,
			format.Simplify(),
		)
		if err != nil {
			return err
		}

		str := string(bytes)
		
		fmt.Println(str)

		*/

	}

	return nil
}
