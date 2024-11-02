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
	"fmt"
	"os"
	"io"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/gdamore/tcell/v2"

	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)

type TextEditor struct {
	*tview.TextView

	W io.Writer

	OnChange func()
}

func NewTextEditor(onchange func()) *TextEditor {

	te := &TextEditor{
		TextView: tview.NewTextView(),
		OnChange: onchange,
	}
	te.SetWordWrap(true).
		SetDynamicColors(true).
		SetBorder(true)
	te.SetChangedFunc(te.OnChange)

	te.W = tview.ANSIWriter(te)

	return te
}

func (ED *TextEditor) OpenFile(path string) {

	body, err := os.ReadFile(path)
	if err != nil {
		tui.SendCustomEvent("/console/err", err.Error())
	}

	l := lexers.Match(path)
	lexer := "text"	
	if l != nil {
		lexer = l.Config().Name
	} else {
		var s string
		if len(body) > 512 {
			s = string(body[:512])
		} else {
			s = string(body)
		}
			
		l = lexers.Analyse(s)
		if l != nil {
			lexer = l.Config().Name
		}
	}

	ED.SetTitle(fmt.Sprintf("%s (%s)", path, lexer))

	ED.Clear()

	err = quick.Highlight(ED.W, string(body), lexer, "terminal256", "github-dark")
	if err != nil {
		tui.SendCustomEvent("/console/err", err.Error())
	}

	ED.Focus(func(p tview.Primitive){})

	ED.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {

		switch evt.Key() {

		case tcell.KeyRune:
			switch evt.Rune() {
				case '?':
				tui.SendCustomEvent("/console/err", err.Error())
				return nil
			default:
				return evt
			}

		default:
			return evt
		}

		// VB.Rebuild("")

		return nil
	})
}

func (ED *TextEditor) Focus(delegate func(p tview.Primitive)) {
	delegate(ED.TextView)
}
