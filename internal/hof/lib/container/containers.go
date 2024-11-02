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

package container

import (
	"context"
	"time"
)

func GetImages(ref string) ([]Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return rt.Images(ctx, Ref(ref))
}

func GetContainers(name string) ([]Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return rt.Containers(ctx, Name(name))
}

func StartContainer(ref, name string, env []string, replace bool) error {
	if replace {
		StopContainer(name)
	}

	return rt.Run(context.Background(), Ref(ref), Params{
		Name: Name(name),
		Env:  env,
	})
}

func StopContainer(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return rt.Remove(ctx, Name(name))
}

func PullImage(ref string) error {
	return rt.Pull(context.Background(), Ref(ref))
}
