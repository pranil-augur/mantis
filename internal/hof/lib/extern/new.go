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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentofu/opentofu/internal/hof/lib/yagu"
)

func NewEntry(what, name, template, strData string) (string, error) {
	if template == "" || template[0] == '#' || template[0] == '@' {
		template = "https://github.com/hofstadter-io/studios-templates" + template
	}

	origName := name

	dir, fn := filepath.Split(name)
	if dir == name {
		dir = name
	}
	name = fn

	// prep data
	var data map[string]interface{}

	if strData != "" {
		var iface interface{}
		var err error

		// is it readable as a file?
		/* TODO fix
		_, err = io.ReadFile(strData, &iface)
		if err == nil {
			data = iface.(map[string]interface{})

		} else {
			// is it readable as string data?
			_, err = io.ReadAll(strings.NewReader(strData), &iface)
			if err == nil {
				data = iface.(map[string]interface{})
			} else {
				// we can't handle the data
				return "", errors.New("supplied <data> is neither a readable file or in a known format as a string")
			}

		}
		*/

		data["name"] = name

	} else {
		// name is the only thing to pass as data
		data = map[string]interface{}{
			"name": name,
		}
	}

	// Prep contextual data
	_, appname := yagu.GetAcctAndName()
	data["AppName"] = appname

	// A bit hacky
	paths := strings.Split(origName, "/")
	if len(paths) >= 3 {
		data["ModuleName"] = paths[2]
	}

	url, version, subpath := SplitParts(template)
	basePath := dir

	switch what {

	case "module":
		basePath = filepath.Join(dir, name)
		data["ModuleName"] = name
		if subpath == "" {
			subpath = "module-default"
		}

	case "type":
		data["TypeName"] = name
		if subpath == "" {
			subpath = "type-default"
		}

	case "page":
		if len(paths) >= 5 {
			data["TypeName"] = paths[3]
		}
		data["PageName"] = name
		if subpath == "" {
			subpath = "page-default"
		}

	case "component":
		if len(paths) >= 5 {
			data["TypeName"] = paths[3]
		}
		data["ComponentName"] = name
		if subpath == "" {
			subpath = "component-default"
		}

	default:
		return "", errors.New("Unknown new what: " + what)
	}

	data["BasePath"] = strings.TrimSuffix(basePath, "/")

	err := cloneAndRenderNewThing(url, version, subpath, basePath, name, data)
	if err != nil {
		return "", err
	}

	// TODO be sure to add the module to your app.modules
	return fmt.Sprintf("Created %s %s", what, origName), nil
}

func cloneAndRenderNewThing(srcUrl, srcVer, srcSubpath, destBasePath, name string, data map[string]interface{}) error {

	// fmt.Printf("%q %q %q %q -> %q\n", name, srcUrl, srcVer, srcSubpath, destBasePath)
	// fmt.Println(data)

	var err error
	var dir string

	if strings.HasPrefix(srcUrl, "https") {
		dir, err = yagu.CloneRepo(srcUrl, srcVer)
		if err != nil {
			return err
		}
	} else {
		// assume local, just copy, so working copy
		dir = srcUrl
	}

	err = yagu.RenderDirNameSub(filepath.Join(dir, srcSubpath, "design"), destBasePath, data)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, srcSubpath, "design-vendor")); !os.IsNotExist(err) {
		// path exists
		err = yagu.RenderDirNameSub(filepath.Join(dir, srcSubpath, "design-vendor"), destBasePath, data)
		if err != nil {
			return err
		}
	}

	return nil
}
