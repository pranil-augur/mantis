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

package extern

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
)

func ImportAddBundle(bundle string) (string, error) {
	if bundle == "" || bundle[0] == '#' || bundle[0] == '@' {
		bundle = "https://github.com/hofstadter-io/studios-modules" + bundle
	}
	url, version, subpath := SplitParts(bundle)

	err := cloneAndRenderImport(url, version, subpath)
	if err != nil {
		return "", err
	}

	// TODO update some deps file

	return "Done", nil
}

func cloneAndRenderImport(srcUrl, srcVer, srcPath string) error {
	_, appname := yagu.GetAcctAndName()
	data := map[string]interface{}{
		"AppName": appname,
	}

	dir, err := yagu.CloneRepo(srcUrl, srcVer)
	if err != nil {
		return err
	}

	err = yagu.RenderDir(filepath.Join(dir, srcPath, "design"), "design-vendor", data)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(dir, srcPath, "design-vendor")); !os.IsNotExist(err) {
		// path exists
		err = yagu.RenderDir(filepath.Join(dir, srcPath, "design-vendor"), "design-vendor", data)
		if err != nil {
			return err
		}
	}
	return nil
}

func SplitParts(full string) (url, version, subpath string) {
	posVersion := strings.LastIndex(full, "@")
	posSubpath := strings.LastIndex(full, "#")

	if posVersion == -1 && posSubpath == -1 {
		url = full
		return
	}

	if posVersion == -1 {
		parts := strings.Split(full, "#")
		url, subpath = parts[0], parts[1]
		return
	}

	if posSubpath == -1 {
		parts := strings.Split(full, "@")
		url, version = parts[0], parts[1]
		return
	}

	if posVersion < posSubpath {
		parts := strings.Split(full, "#")
		subpath = parts[1]
		parts = strings.Split(parts[0], "@")
		url, version = parts[0], parts[1]
	} else {
		parts := strings.Split(full, "@")
		version = parts[1]
		parts = strings.Split(parts[0], "#")
		url, subpath = parts[0], parts[1]
	}

	return url, version, subpath
}
