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

package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/tui"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/app"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/events"

	"github.com/opentofu/opentofu/internal/hof/lib/tui/modules"
)

func Cmd(args []string, rflags flags.RootPflagpole) error {

	// setup new app 
	App, err := app.NewApp()
	if err != nil {
		return err
	}

	// catch panics and exit
	defer func() {
		err := recover()
		if err != nil {
			App.Stop()
			time.Sleep(time.Second)
			fmt.Println("PANIC'D HERE.tui.Cmd")
			panic(err)
		}
	}()

	tui.SetApp(App)

	// initialize our modules
	modules.Init()

	// Set the root view
	root := modules.RootView()
	App.SetRootView(root)

	// Ctrl-c to quit program
	tui.AddGlobalHandler("/sys/key/C-A-c", func(e events.Event) {
		App.Stop()
	})

	// Log Key presses (if you want to)
	if tkl := os.Getenv("HOF_TUI_KEYLOGGER"); tkl != "" {
		tklb, _ := strconv.ParseBool(tkl)
		if tklb {
			logKeys()
		}
	}

	// Run PProf (useful for catching hangs)
	// go runPprofServer()

	// fmt.Printf("tui.Cmd args: %v\n", args)
	tui.Log("trace", fmt.Sprintf("tui.Cmd args: %v", args))

	// some special cases to deal with CLI base startup
	path := "eval"
	if len(args) == 0 {
		args = []string{"eval", "help"}
	} else {
		switch args[0] {
		case "eval":
			if len(args) == 1 {
				args = []string{"eval", "tree"}
			}
		case "play":
			if len(args) == 1 {
				args = []string{"eval", "play", "new"}
			}
		}
	}

	context := map[string]any{
		"page": path,
		"args": args,
	}

	tui.Log("trace", fmt.Sprintf("tui.Cmd context: %# v", context))

	go func() {
		// some latent locksups occur randomly
		time.Sleep(time.Millisecond * 23)
		tui.SendCustomEvent("/status/message", "[dodgerblue::b]Welcome to [gold::bi]_[ivory]Hofstadter[-::-]")
	}()

	// Start the Main (Blocking) Loop
	return App.Start(context)
}

func logKeys() {
	tui.AddGlobalHandler("/sys/key", func(e events.Event) {
		if k, ok := e.Data.(events.EventKey); ok {
			go tui.SendCustomEvent("/console/info", "key: " + k.KeyStr)
		}
	})
	tui.AddGlobalHandler("/sys/mouse", func(e events.Event) {
		if k, ok := e.Data.(events.EventMouse); ok {
			b := k.Buttons()
			if 0 < b && b < 256 {
				go tui.SendCustomEvent("/console/info", fmt.Sprintf("key: %d", k.Buttons()))
			}
		}
	})
}

func runPprofServer() {
	runtime.SetMutexProfileFraction(1)
	http.ListenAndServe(":8888", nil)
}
