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

package cuetils

import (
	"os"
	"strings"

	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/load"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
)

// CueRuntimeFromArgs builds up a CueRuntime
//  by processing the args passed in
func CueRuntimeFromEntrypoints(entrypoints []string) (crt *CueRuntime, err error) {
	crt = &CueRuntime{
		Entrypoints: entrypoints,
		CueConfig: &load.Config{
			ModuleRoot: "",
			Module:     "",
			Package:    "",
			Dir:        "",
			Tags:       []string{},
			TagVars:    load.DefaultTagVars(),
			Tests:      false,
			Tools:      false,
			DataFiles:  false,
			Overlay:    map[string]load.Source{},
		},
		dataMappings: make(map[string]string),
	}

	err = crt.Load()

	return crt, err
}

// CueRuntimeFromArgsAndFlags builds up a CueRuntime
//  by processing the args passed in AND the current flag values
func CueRuntimeFromEntrypointsAndFlags(entrypoints []string) (crt *CueRuntime, err error) {
	rflags := flags.RootPflags
	cfg := &load.Config{
		ModuleRoot: "",
		Module:     "",
		Package:    "",
		Dir:        "",
		Tags:       rflags.Tags,
		TagVars:    load.DefaultTagVars(),
		Tests:      false,
		Tools:      false,
		DataFiles:  false,
		Overlay:    make(map[string]load.Source),
	}

	// package?
	if rflags.Package != "" {
		cfg.Package = rflags.Package
	}
	if rflags.InjectEnv {
		for _, e := range os.Environ() {
			parts := strings.Split(e, "=")
			k,v := parts[0], parts[1]
			cfg.TagVars[k] = load.TagVar{
				Func: func() (ast.Expr, error) {
					return ast.NewString(v), nil
				},
			}
		}
	}

	crt = &CueRuntime{
		Entrypoints: entrypoints,
		CueConfig:   cfg,
		IncludeData: rflags.IncludeData,
		dataMappings: make(map[string]string),
	}

	err = crt.Load()

	return crt, err
}

