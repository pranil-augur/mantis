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

package yagu

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

// SHA1, SHA256, MD5
const (
	SHA1   = "sha1"
	SHA256 = "sha256"
	MD5    = "md5"
)

// NewHash create an instance of specific hash algorithm
func NewHash(name *string) (hash.Hash, error) {
	if *name == SHA1 {
		return sha1.New(), nil
	} else if *name == SHA256 {
		return sha256.New(), nil
	} else if *name == MD5 {
		return md5.New(), nil
	}
	message := "Hash Algorithm is not supported"
	err := errors.New(message)
	return nil, err

}

// Create hash value with local path and a hash algorithm
func GetHash(path string, hashAlgorithm string) (string, error) {

	hash, err := NewHash(&hashAlgorithm)

	if err != nil {
		return "", nil
	}

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		io.WriteString(hash, path)
		return nil
	})

	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
