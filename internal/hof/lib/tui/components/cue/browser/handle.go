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

package browser

import (
	"fmt"

	"github.com/opentofu/opentofu/internal/hof/lib/tui"
)

// local handler type
type HandlerFunc func (B *Browser, action string, args []string, context map[string]any) (handled bool, err error)

// action registry
var actions = map[string]HandlerFunc{
	"clear":  handleClear,
	"create": handleSet,
	"set":    handleSet,
	"add":    handleAdd,
	"conn":   handleAdd,
	"watch":  handleWatchConfig,
	"globs":  handleWatchConfig,
}

// implementation of widget.ActionHandler interface
func (B *Browser) HandleAction(action string, args []string, context map[string]any) (bool, error) {
	tui.Log("warn", fmt.Sprintf("Browser.HandleAction: %v %v %v", action, args, context))

	handler, ok := actions[action]
	if ok {
		return handler(B, action, args, context)
	}

	return false, nil
}
