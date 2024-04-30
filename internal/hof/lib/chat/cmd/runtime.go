/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package cmd

import (
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	"github.com/opentofu/opentofu/internal/hof/lib/chat"
	"github.com/opentofu/opentofu/internal/hof/lib/runtime"
)

func prepRuntime(args []string, rflags flags.RootPflagpole) (*runtime.Runtime, error) {

	// create our core runtime
	r, err := runtime.New(args, rflags)
	if err != nil {
		return nil, err
	}

	err = r.Load()
	if err != nil {
		return nil, err
	}

	err = r.EnrichChats(nil, EnrichChat)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func EnrichChat(R *runtime.Runtime, c *chat.Chat) error {

	// no-op
	return nil
}
