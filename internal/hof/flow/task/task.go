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

package task

import (
	"time"

	"cuelang.org/go/cue"
	cueflow "cuelang.org/go/tools/flow"
	"github.com/google/uuid"

	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

type Task interface {
	IDer
	Eventer
	TimeEventer
}

type IDer interface {
	ID() string
	UUID() string
}

type Eventer interface {
	EmitEvent(key string, data interface{})
}

type TimeEventer interface {
	AddTimeEvent(key string)
}

type BaseTask struct {
	// IDer
	ID   string
	UUID uuid.UUID

	// cue bookkeeping
	CueTask *cueflow.Task
	Node    *hof.Node[any]
	Orig    cue.Value
	Start   cue.Value
	Final   cue.Value
	Error   error

	// stats & timing
	// should this be a list with names / times
	// timing
	// replace with open telemetry
	TimeEvents map[string]time.Time
}

func NewBaseTask(node *hof.Node[any]) *BaseTask {
	val := node.Value
	return &BaseTask{
		ID:         val.Path().String(),
		UUID:       uuid.New(),
		Node:       node,
		Orig:       val,
		TimeEvents: make(map[string]time.Time),
	}
}

func (T *BaseTask) AddTimeEvent(key string) {
	T.TimeEvents[key] = time.Now()
}
