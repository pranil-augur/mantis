/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package chat

import (
	"github.com/pkoukk/tiktoken-go"
)

// https://github.com/pkoukk/tiktoken-go
// https://pkg.go.dev/github.com/pkoukk/tiktoken-go

// count tokens so we can inform users and better size our requests

var _ = tiktoken.Tiktoken{}


