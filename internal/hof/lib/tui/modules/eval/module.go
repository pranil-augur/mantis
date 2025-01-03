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

package eval

import (
	"github.com/opentofu/opentofu/internal/hof/lib/connector"
)

func New() connector.Connector {
	items := []any{
		NewEval(),
	}

	m := connector.New("Eval")
	m.Add(items)

	return m
}

func (M *Eval) Id() string {
	return "eval"
}

func (M *Eval) Name() string {
	return "Eval"
}

func (M *Eval) HotKey() string {
	return ""
}

func (M *Eval) CommandName() string {
	return "eval"
}

func (M *Eval) CommandUsage() string {
	return "eval"
}

func (M *Eval) CommandHelp() string {
	return "help for eval module"
}

// CommandCallback is invoked when the user runs your module
// your goal is to enrich the context with the args
// return the object you want in Refresh
func (M *Eval) CommandCallback(context map[string]any) {
	// tui.Log("extra", fmt.Sprintf("Eval.CmdCallback: %# v", context))
	context = enrichContext(context)
	args := []string{}
	if _args, ok := context["args"]; ok {
		args = _args.([]string)
	}

	// handle any top-leval eval commands
	action := ""
	if _action, ok := context["action"]; ok {
		action = _action.(string)
	}

	M.HandleAction(action, args, context)

	return
}
