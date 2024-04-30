/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package runtime

import (
	"fmt"
	"time"

	"github.com/opentofu/opentofu/internal/hof/lib/chat"
)

type ChatEnricher func(*Runtime, *chat.Chat) error

func (R *Runtime) EnrichChats(chats []string, enrich ChatEnricher) error {
	start := time.Now()
	defer func() {
		end := time.Now()
		R.Stats.Add("enrich/chat", end.Sub(start))
	}()

	if R.Flags.Verbosity > 1 {
		fmt.Println("Runtime.Chat: ", chats)
		for _, node := range R.Nodes {
			node.Print()
		}
	}

	// Find only the datamodel nodes
	// TODO, dedup any references
	cs := []*chat.Chat{}
	for _, node := range R.Nodes {
		// check for DM root
		if node.Hof.Chat.Root {

			cs = append(cs, &chat.Chat{Node: node})
		}
	}

	R.Chats = cs

	for _, c := range R.Chats {
		err := enrich(R, c)
		if err != nil {
			return err
		}
	}


	return nil
}
