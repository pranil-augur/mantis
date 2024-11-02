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
 * This work is based on code from https://github.com/gofireflyio/aiac, licensed under the MIT License.
 */
package openai

import (
	"context"
	"fmt"
	"sort"

	"github.com/opentofu/opentofu/internal/hof/lib/codegen/types"
)

// ListModels returns a list of all the models supported by this backend.
func (backend *OpenAI) ListModels(ctx context.Context) (
	models []string,
	err error,
) {
	var answer struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	err = backend.
		NewRequest("GET", "/models").
		Into(&answer).
		RunContext(ctx)
	if err != nil {
		return models, fmt.Errorf("failed sending prompt: %w", err)
	}

	if len(answer.Data) == 0 {
		return models, types.ErrNoResults
	}

	models = make([]string, len(answer.Data))
	for i := range answer.Data {
		models[i] = answer.Data[i].ID
	}

	sort.Strings(models)

	return models, nil
}
