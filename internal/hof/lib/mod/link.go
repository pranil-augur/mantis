/*
 * Augur AI Proprietary
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
