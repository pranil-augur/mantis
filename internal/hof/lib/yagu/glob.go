/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package yagu

import (
	"sort"

	// TODO, be consistent
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mattn/go-zglob"
)

func CheckShouldInclude(filename string, includes, excludes []string) (bool, error) {
	var err error
	include := false

	if len(includes) > 0 {
		for _, pattern := range includes {
			include, err = doublestar.PathMatch(pattern, filename)
			if err != nil {
				return false, err
			}
			if include {
				break
			}
		}
	} else {
		include = true
	}

	exclude := false
	if len(excludes) > 0 {
		for _, pattern := range excludes {
			exclude, err = doublestar.PathMatch(pattern, filename)
			if err != nil {
				return false, err
			}
			if exclude {
				break
			}
		}
	}

	return include && !exclude, nil
}

func FilesFromGlobs(patterns []string) ([]string, error) {
	// get glob matches
	files := []string{}
	for _, pattern := range patterns {
		matches, err := zglob.Glob(pattern)
		if err != nil {
			return nil, err
		}
		files = append(files, matches...)
	}

	// make unique
	keys := make(map[string]bool)
	unique := make([]string, 0, len(files))
	for _, file := range files {
		if _, value := keys[file]; !value {
			keys[file] = true
			unique = append(unique, file)
		}
	}

	// also sort
	sort.Strings(unique)
	return unique, nil
}

