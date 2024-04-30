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
	"fmt"
	"strings"
)

func ValidateModURL(mod string) error {
	parts := strings.Split(mod, "/")	
	if len(parts) < 2 {
		return fmt.Errorf("error: modules require one or more '/', you provided %q", mod)
	}
	if !strings.Contains(parts[0], ".") {
		return fmt.Errorf("error: the first part of a module path must be a domain, you provided %q", mod)
	}

	return nil
}
