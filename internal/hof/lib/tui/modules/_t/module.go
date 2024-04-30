/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package {{ .name | lower }}

import "github.com/opentofu/opentofu/internal/hof/lib/connector"

func New() connector.Connector {
	items := []any{
		New{{ .name | title }}(),
	}
	m := connector.New("{{ .name | title }}")
	m.Add(items)

	return m
}

