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
	"os"
	"path/filepath"
	"strings"

	// gomod "golang.org/x/mod/module"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/verinfo"
)

var initFileContent = `module: %q
cue: %q
`

func Init(module string, rflags flags.RootPflagpole) (err error) {
	upgradeHofMods()

	/*
	err := gomod.CheckPath(module)
	if err != nil {
		return fmt.Errorf("bad module name %q, should have domain format 'domain.com/...'", module)
	}
	*/

	err = ValidateModURL(module)
	if err != nil {
		return err
	}

	_, err = os.Lstat("cue.mod")
	if err != nil {
		if _, ok := err.(*os.PathError); !ok && (strings.Contains(err.Error(), "file does not exist") || strings.Contains(err.Error(), "no such file")) {
			return err
		}
	} else {
		return fmt.Errorf("CUE module already exists in this directory")
	}

	s := fmt.Sprintf(initFileContent, module, verinfo.CueVersion)

	// mkdir & write file
	err = os.MkdirAll(filepath.Join("cue.mod", "pkg"), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("cue.mod/module.cue"), []byte(s), 0644)
	if err != nil {
		return err
	}

	return nil
}

