/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package widget

import (
	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/lib/tui/tview"
)

// base and wrapped tview widgets, temporarily here

// Widget is designed to fit in containers and be serializable
type Widget interface {
	tview.Primitive

	TypeName() string

	Encode() (map[string]any, error)
	Decode(map[string]any) (Widget, error)

	// UpdateValue()
}

type ConnectionReciever interface {
	SetConnection(args []string, sourceGetter func() cue.Value)
}

type ValueProducer interface {
	// function which returns a value
	GetValue() cue.Value

	// wrapper that reuses the path
	GetValueExpr(expr string) func() cue.Value
}

type ActionHandler interface {
	// TODO, Autocomplete
	// ActionList() []string

	HandleAction(action string, args []string, context map[string]any) (bool, error)
}
