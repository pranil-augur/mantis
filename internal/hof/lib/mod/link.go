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

package mod

import (
	"fmt"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
)

func Link(rflags flags.RootPflagpole) (error) {
	upgradeHofMods()

	cm, err := loadRootMod()
	if err != nil {
		return err
	}

	if rflags.Verbosity > 0 {
		fmt.Println("linking deps for:", cm.Module)
	}

	err = cm.ensureCached()
	if err != nil {
		return err
	}

	return cm.Vendor("link", rflags.Verbosity)
}
