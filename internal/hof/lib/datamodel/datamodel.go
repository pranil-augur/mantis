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

package datamodel

import (
	"github.com/opentofu/opentofu/internal/hof/lib/hof"
)

// this is basically the same as a Value
// except that it reporesents a conceptual root
// and we want specific functions on it
// that are different from general Nodes'
// handling and recursion
// type Datamodel *hof.Node[Value]
type Datamodel struct {
	*hof.Node[Value]
}

func DatamodelType(DM *Datamodel) string {
	// if explicitly set to CUE value
	//   todo, can we look for incomplete values?
	//   seems problematic, can't separate bad config
	//   so probably not, but leaving this comment here
	if DM.Hof.Datamodel.Cue {
		return "value"
	}

	// if history at root & no children ... a bit hacky, but will do
	//   is Root correct here? what about just hist & no children? (all leaf nodes are then objects?
	// if DM.Hof.Datamodel.Root && DM.Hof.Datamodel.History && len(DM.Children) == 0 {
	if DM.Hof.Datamodel.History && len(DM.Children) == 0 {
		return "object"
	}

	// otherwise generic datamodel
	return "datamodel"
}

func (dm *Datamodel) Status() string {
	if has, _ := dm.HasHistory(); !has {
		return "no-history"
	}

	// TODO... dirty or version
	if dm.HasDiff() {
		return "dirty"
	}

	return "ok"
}
