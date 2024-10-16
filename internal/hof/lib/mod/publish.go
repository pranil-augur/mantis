/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package mod

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/opentofu/opentofu/internal/hof/lib/repos/remote"
)

func Publish(taggedMod string) error {
	parts := strings.Split(taggedMod, ":")

	var (
		mod string
		tag string
	)

	switch {
	case len(parts) == 1:
		mod = taggedMod
		tag = "latest"
	case len(parts) == 2:
		mod = parts[0]
		tag = parts[1]
	default:
		return errors.New("invalid mod")
	}

	taggedMod = fmt.Sprintf("%s:%s", mod, tag)

	rmt, err := remote.Parse(mod)
	if err != nil {
		return fmt.Errorf("remote parse: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os get wd: %w", err)
	}

	// d := filepath.Join(wd, "cue.mod", "pkg", mod)

	ctx := context.Background()
	if err = rmt.Publish(ctx, wd, taggedMod); err != nil {
		return fmt.Errorf("remote publish: %w", err)
	}

	return nil
}
