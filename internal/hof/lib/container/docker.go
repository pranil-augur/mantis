/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package container

type docker struct {
	runtime
}

func newDocker() docker {
	return docker{
		runtime: newRuntime(RuntimeBinaryDocker),
	}
}
