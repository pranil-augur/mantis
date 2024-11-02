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

{{ $Name := .name | title -}}
{{ $name := .name | lower -}}
package {{ $name }}

import (
	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/hoc/router"
)

type {{ $Name }} struct {
	*tview.Flex
}

func New{{ $Name }}() *{{ $Name }} {
	m := &{{ $Name }}{
		Flex: tview.NewFlex(),
	}

	// do layout setup here

	return m
}

func (M *{{ $Name }}) Id() string {
	return "{{ $name }}"
}

func (M *{{ $Name }}) Routes() []router.RoutePair {
	return []router.RoutePair{
		router.RoutePair{"/{{ $name }}", M},
	}
}

func (M *{{ $Name }}) Name() string {
	return "{{ $Name }}"
}

func (M *{{ $Name }}) HotKey() string {
	return ""
}

func (M *{{ $Name }}) CommandName() string {
	return "{{ $name }}"
}

func (M *{{ $Name }}) CommandUsage() string {
	return "{{ $name }}"
}

func (M *{{ $Name }}) CommandHelp() string {
	return "help for {{ $name }} module"
}

// CommandCallback is invoked when the user runs your module
// return the object you want in mount or refresh
func (M *{{ $Name }}) CommandCallback(args []string, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]any)
	}
	context["args"] = args

	if M.IsMounted() {
		// just refresh with new args
		M.Refresh(context)
	} else {
		// need to navigate, mount will do the rest
		context["path"] = "/{{ $name }}"
		go tui.SendCustomEvent("/router/dispatch", context)
	}
}

func (M *{{ $Name }}) Mount(context map[string]any) error {
	// this is where we can do some loading
	M.Flex.Mount(context)

	err := M.Refresh(context)
	if err != nil {
		tui.SendCustomEvent("/console/error", err)
		return err
	}

	// mount any other components
	// maybe we should have [...Children], so two pointers, one for dev, one for sys (Children)
	// then this call to mount can be handled without extra stuff by default?
	//M.View.Mount(context)
	//M.Eval.Mount(context)


	return nil
}

func (M *{{ $Name }}) Unmount() error {
	// this is where we can do some unloading, depending on the application
	//M.View.Unmount()
	//M.Eval.Unmount()
	M.Flex.Unmount()

	return nil
}

func (M *{{ $Name }}) Refresh(context map[string]any) error {

	// this is where you update data and set in components
	// then at the end call tui.Draw()

	return nil
}
