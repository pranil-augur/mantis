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

import (
	"context"
	"fmt"
	"time"
)

type RuntimeVersion struct {
	Name string
	Client struct {
		Version    string
		APIVersion string
	}
	Server struct {
		Version       string
		APIVersion    string
		MinAPIVersion string
	}
}

func GetBinary() (string) {
	return rt.Binary()
}

func GetVersion() (RuntimeVersion, error) {
	if rt.Binary() == "none" {
		return RuntimeVersion{Name: "none"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return rt.Version(ctx)
}

func (r runtime) Version(ctx context.Context) (RuntimeVersion, error) {
	var rv RuntimeVersion
	if err := r.execJSON(ctx, &rv, "version", "--format", "{{ json . }}"); err != nil {
		return rv, fmt.Errorf("exec json: %w", err)
	}

	rv.Name = string(r.bin)
	return rv, nil
}

func (r RuntimeVersion) String() string {
	return fmt.Sprintf(
		"%s [%s (client) | %s (server)]",
		r.Name,
		r.Client.Version,
		r.Server.Version,
	)
}
