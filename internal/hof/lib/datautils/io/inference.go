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

package io

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/naoina/toml"
	"gopkg.in/yaml.v2"
)

/*
Where's your docs doc?!
*/
func InferDataContentType(data []byte) (contentType string, err error) {

	// TODO: look for unique symbols in the data
	// but always try to unmarshal to be sure

	var obj interface{}

	err = json.Unmarshal(data, &obj)
	if err == nil {
		return "json", nil
	}

	err = yaml.Unmarshal(data, &obj)
	if err == nil {
		return "yaml", nil
	}

	err = xml.Unmarshal(data, &obj)
	if err == nil {
		return "yaml", nil
	}

	err = toml.Unmarshal(data, &obj)
	if err == nil {
		return "toml", nil
	}

	return "", errors.New("[IDCT] unknown content type")

	return
}

/*
Where's your docs doc?!
*/
func InferFileContentType(filename string) (contentType string, err error) {

	// assume files have correct extensions
	// TODO use 'filepath.Ext()'
	ext := filepath.Ext(filename)[1:]
	switch ext {

	case "json":
		return "json", nil

	case "toml":
		return "toml", nil

	case "yaml", "yml":
		return "yaml", nil

	case "xml":
		return "xml", nil

	default:
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return InferDataContentType(data)
	}

	return
}

// HOFSTADTER_BELOW
