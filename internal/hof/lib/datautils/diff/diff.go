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

package diff

import (
	"fmt"

	"github.com/pkg/errors"
)

func Diff(original interface{}, current interface{}) (diff interface{}, err error) {

	fmt.Println("DIFF'n types: " + fmt.Sprintf("%T, %T", original, current))
	// check that they are the same type at the root
	// If different - error
	// If same - go recurse
	switch original.(type) {

	case map[string]interface{}:
		_, ok := current.(map[string]interface{})
		if !ok {
			return nil, errors.New("undiffable types, not the same type" + fmt.Sprintf("%T, %T", original, current))
		}

		return rdiff(original, current)

	case []interface{}:
		_, ok := current.([]interface{})
		if !ok {
			return nil, errors.New("undiffable types, not the same type" + fmt.Sprintf("%T, %T", original, current))
		}

		return rdiff(original, current)

	default:
		// TODO check for golang types with reflect
		return nil, errors.New("undiffable original, must be map or slice" + fmt.Sprintf("%T, %+v", original, original))

	}

	return nil, errors.New("undiffable original" + fmt.Sprintf("%T, %+v", original, original))
}

func rdiff(original interface{}, current interface{}) (diff interface{}, err error) {

	switch O := original.(type) {

	case map[string]interface{}:
		C, ok := current.(map[string]interface{})
		if !ok {
			return nil, errors.New("undiffable types, not the same type" + fmt.Sprintf("%T, %T", original, current))
		}

		fmt.Println("diffable types" + fmt.Sprintf("%T, %T", O, C))

		return nil, nil

	case []interface{}:
		C, ok := current.([]interface{})
		if !ok {
			return nil, errors.New("undiffable types, not the same type" + fmt.Sprintf("%T, %T", original, current))
		}

		fmt.Println("diffable types" + fmt.Sprintf("%T, %T", O, C))

		// if elements have names,

		return nil, nil

	default:
		// TODO check for golang types with reflect
		return nil, errors.New("undiffable original, must be map or slice" + fmt.Sprintf("%T, %+v", original, original))

	}

	return nil, errors.New("undiffable known" + fmt.Sprintf("%#+v, %#+v", original, current))
}
