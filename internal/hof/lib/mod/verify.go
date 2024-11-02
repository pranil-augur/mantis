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
	"github.com/opentofu/opentofu/internal/hof/lib/repos/cache"
)

func Verify(rflags flags.RootPflagpole) (error) {
	upgradeHofMods()

	cm, err := loadRootMod()
	if err != nil {
		return err
	}

	if len(cm.Require) == 0 {
		fmt.Println("no requirements found")
		return nil
	}

	err = cm.ensureCached()
	if err != nil {
		return err
	}

	return cm.Verify()
}

func (cm *CueMod) Verify() (err error) {

	for path, ver := range cm.Require {
		if _, ok := cm.Replace[path]; !ok {
			err := cm.verifyModule(path, ver)
			if err != nil {
				return err
			}
		}
	}

	for path, ver := range cm.Indirect {
		if _, ok := cm.Replace[path]; !ok {
			err := cm.verifyModule(path, ver)
			if err != nil {
				return err
			}
		}
	}

	return nil
}


func (cm *CueMod) verifyModule(path, ver string) (error) {

	// get hash from cache
	hc, err := cache.Checksum(path, ver)
	if err != nil {
		return err
	}

	// get has from sums
	hs, ok := cm.Sums[Dep{path,ver}]
	if !ok {
		return fmt.Errorf("%s@%s missing from sum file", path, ver)
	}

	// search
	match := false
	for _, h := range hs {
		if h == hc {
			match = true
			break
		}
	}

	if !match {
		return fmt.Errorf(mismatchMsg, path, ver, []string{hc}, hs)
	}

	return nil
}

var mismatchMsg = `unable to verify %s@%s incorrect or missing hashes
    cache:    %v
    sums.cue: %v
`
