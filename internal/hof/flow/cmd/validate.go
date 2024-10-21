/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 */

package cmd

import "github.com/opentofu/opentofu/internal/hof/lib/mantis"

func Validate(dir string) error {
	return mantis.Validate(dir)
}
