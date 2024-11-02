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

package common

import (
	"github.com/opentofu/opentofu/internal/hof/lib/tui/app"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)

type Shell struct {
	*tview.TextArea

	text string

	App *app.App
}

func NewShell(app *app.App) *Shell {
	s := &Shell{
		TextArea: tview.NewTextArea(),
		App: app,
	}

	// lower-level setup
	s.SetTitle("Shell").
		SetBorder(true)

	return s
}

func (S *Shell) Write(text string) {
	S.text = text
	S.SetText(S.text, true)
}

func (S *Shell) Append(text string) {
	S.text += text
	S.SetText(S.text, true)
}
