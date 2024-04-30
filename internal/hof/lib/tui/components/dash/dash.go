/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package dash

import (
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)

type Dash struct {
	*tview.Frame

	Flex *tview.Flex

}

func New() *Dash {
	return &Dash{
		Frame: tview.NewFrame(),
		Flex:  tview.NewFlex(),
	}
}

func (C *Dash) Focus(delegate func(p tview.Primitive)) {
	// this is where you can choose how to focus

	// we just delegate to the Flex
	delegate(C.Flex)
}

func (C *Dash) Mount(context map[string]any) error {

	C.Flex.Mount(context)

	// do any setup, mount any subcomponents

	return nil
}

func (C *Dash) Unmount() error {

	// do any teardown, unmount any subcomponents

	C.Flex.Unmount()

	return nil
}
