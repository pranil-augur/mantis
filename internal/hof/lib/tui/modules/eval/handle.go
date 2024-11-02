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
	"fmt"
	"strings"

	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/components/panel"
)

func (M *Eval) HandleAction(action string, args []string, context map[string]any) (bool, error) {

	// default action to update if item is set
	if action == "" {
		// default action to update if item is set
		if _, ok := context["item"]; ok {
			context["action"] = "update"
		} else {
			err := fmt.Errorf("unknown command: %q {%s}", action, strings.Join(args, ","))
			tui.Tell("error", err)
			tui.Log("error", err)
			return true, nil
		}	
	}

	var p *panel.Panel

	// was a specific panel target set?
	if _path, ok := context["target-panel"]; ok {
		path := _path.(string)
		srcItem, err := M.getPanelByPath(path)
		if err != nil {
			return true, err
		}
		if srcItem != nil {
			p = srcItem
		} else {
		  return true, fmt.Errorf("target %q is not a panel", path)
		}
	} else {
		// get the current focused panel
		p = M.GetMostFocusedPanel()
		if p == nil {
			p = M.Panel
		}
	}

	// handle dashboard (panel collection), CRUD actions
	handled, err := M.handleDashActions(p, action, args, context)
	if err != nil {
		tui.Tell("error", err)
		tui.Log("error", err)
		return true, err
	}
	if handled {
		tui.Log("warn", fmt.Sprintf("Eval.handleDashActions: %v %v %v", handled, action, args))
		return true, nil
	}

	// panel actions (that need to happen at the eval level)
	handled, err = M.handlePanelActions(p, action, args, context)
	if handled {
		tui.Log("warn", fmt.Sprintf("Eval.handlePanelActions: %v %v", action, args))
		p.Refresh(context)
		return true, nil
	}

	// item actions (passed down to current item)
	handled, err = M.handleItemActions(p, action, args, context)
	if handled {
		tui.Log("warn", fmt.Sprintf("Eval.handleItemActions: %v %v", action, args))
		p.Refresh(context)
		return true, nil
	}

	if !handled {
		err = fmt.Errorf("unhandled inputs: %v %v", action, args)
		tui.Tell("crit", err)
		tui.Log("error", err)
	}

	p.Refresh(context)


	return false, fmt.Errorf("unhandled action %s %v %#v", action, args, context)
}
