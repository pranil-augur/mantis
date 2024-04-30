/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package modules

import (
	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"

	"github.com/opentofu/opentofu/internal/hof/lib/connector"

	// base modules
	"github.com/opentofu/opentofu/internal/hof/lib/tui/modules/root"
	"github.com/opentofu/opentofu/internal/hof/lib/tui/modules/help"

	// core modules
	"github.com/opentofu/opentofu/internal/hof/lib/tui/modules/eval"

	// extra modules
	"github.com/opentofu/opentofu/internal/hof/lib/tui/modules/ls"

)

var (
	Conn   connector.Connector
	rootView tview.Primitive
)

func Init() {
	rootView = root.New()

	items := []interface{}{
		// primary / root layout component
		rootView,

		// base modules
		help.New(),

		// core modules
		eval.New(),

		// extra modules
		ls.New(),
	}

	conn := connector.New("root")
	conn.Add(items)
	Conn = conn

	Conn.Connect(Conn)
}

func RootView() tview.Primitive {
	return rootView
}
