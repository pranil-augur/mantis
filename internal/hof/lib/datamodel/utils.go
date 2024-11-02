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
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
)

func (dm *Datamodel) FindMaxLabelLen(dflags flags.DatamodelPflagpole) int {
	max := len(dm.Hof.Label)
	m := dm.T.findMaxLabelLenR("", "  ", dflags)
	if m > max {
		max = m
	}
	return max
}

func (V *Value) findMaxLabelLenR(indent, spaces string, dflags flags.DatamodelPflagpole) int {
	max := V.findMaxLabelLen(indent, spaces, dflags)
	for _, c := range V.Children {
		m := c.T.findMaxLabelLenR(indent + spaces, spaces, dflags)
		if m > max {
			max = m
		}
	}
	return max
}

func (V *Value) findMaxLabelLen(indent, spaces string, dflags flags.DatamodelPflagpole) int {
	return len(V.Hof.Label) + len(indent)
}
