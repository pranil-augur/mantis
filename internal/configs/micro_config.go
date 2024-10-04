// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configs

// MicroConfig holds configuration data with its identifier, content in bytes, and format.
type MicroConfig struct {
	Identifier       string
	Content          []byte
	Format           string // Expected values: "cue", "json", "hcl"
	BackendStatePath string
}
