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
package components

import (
	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)


type {{ $Name }} struct {
	*tview.Flex // or whatever you want

	// whatever else you want
}

func New{{ $Name }}(/* ... */) *{{ $Name }} {
	c := &{{ $Name }} {
		Flex: tview.NewFlex(),
	}

	// other first time / layout setup

	return c
}

func (C *{{ $Name }}) Focus(delegate func(p tview.Primitive)) {
	delegate(C.Flex)
	// this is where you can choose how to focus
}

func (C *{{ $Name }}) Mount(context map[string]any) error {

	C.Flex.Mount(context)

	// do any setup, mount any subcomponents

	return nil
}

func (C *{{ $Name }}) Unmount() error {

	// do any teardown, unmount any subcomponents

	C.Flex.Unmount()

	return nil
}
