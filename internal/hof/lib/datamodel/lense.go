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

import "cuelang.org/go/cue"

type Lense struct {
	// Explination for the snapshot or changes therein
	Description string

	CurrVersion string
	PrevVersion string

	// calculated by hof
	CurrDiff cue.Value  // prev -> this
	PrevDiff cue.Value  // this -> prev

	// user defined transform to cover the gaps
	// make use of structural to support @st(...)
	CurrUser cue.Value  // prev -> this
	PrevUser cue.Value  // this -> prev

	// the full transformation or migration to implement the lens
	// calculated as NextDiff & NextUser => NextMig
	CurrMig cue.Value   // prev -> this
	PrevMig cue.Value   // this -> prev

	// we should have commands to show the above "math"
	// and also process data through versions
}
