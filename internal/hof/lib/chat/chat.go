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

package chat

import (
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type Chat struct {
	*hof.Node[any]

	Name string
	HumanName   string
	MachineName string

	Description        string
	HumanDescription   string
	MachineDescription string

	// user inputs
	Args []string
	Files map[string]string
  Question string

	Model   string

	System   string
	Examples []Example
	Messages []Message
	Parameters map[string]any

}

type Session struct {
	System string
	Examples []Example
	Messages []Message
}

type Example struct {
	Input  string
	Output string
}

type Message struct {
	Role    string
	Content string
}
