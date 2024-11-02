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

package console

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/events"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)

type ErrConsoleWidget struct {
	*tview.TextView
}

func NewErrConsoleWidget() *ErrConsoleWidget {
	textView := tview.NewTextView()
	textView.
		SetTextColor(tcell.ColorMaroon).
		SetScrollable(true).
		SetChangedFunc(func() {
			tui.Draw()
			textView.ScrollToEnd()
		})

	textView.SetTitle(" errors ").
		SetBorder(true).
		SetBorderColor(tcell.ColorRed)

	C := &ErrConsoleWidget{
		TextView: textView,
	}

	return C
}

func (C *ErrConsoleWidget) Mount(context map[string]interface{}) error {

	tui.AddGlobalHandler("/user/error", func(evt events.Event) {
		str := evt.Data.(*events.EventCustom).Data()
		text := fmt.Sprintf("[%s] %v\n", evt.When().Format("2006-01-02 15:04:05"), str)
		fmt.Fprintf(C, "%s", text)
	})

	tui.AddGlobalHandler("/sys/err", func(ev events.Event) {
		err := ev.Data.(*events.EventError)
		line := fmt.Sprintf("[%s] %v", ev.When().Format("2006-01-02 15:04:05"), err)
		fmt.Fprintf(C, "[red]SYSERR %v[white]\n", line)
	})

	return nil
}
func (C *ErrConsoleWidget) Unmount() error {
	tui.RemoveWidgetHandler(C, "/user/error")
	tui.RemoveWidgetHandler(C, "/sys/err")
	return nil
}
