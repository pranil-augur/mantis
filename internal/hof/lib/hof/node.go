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
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package hof

import (
	"cuelang.org/go/cue"
)

type Node[T any] struct {
	Hof Hof

	// do not modify, root value containing
	Value cue.Value

	// The wrapping type
	T *T

	// heirarchy of tracked values
	Parent   *Node[T]
	// we (this node) are in between
	Children []*Node[T]

	// cue paths to get up/down hierarchy
}

func New[T any](label string, val cue.Value, curr *T, parent *Node[T]) *Node[T] {
	n := &Node[T]{
		Hof: Hof{
			Path: val.Path().String(),
			Label: label,
		},
		Value:    val,
		T:        curr,
		Parent:   parent,
		Children: make([]*Node[T], 0),
	}

	return n
}

func Upgrade[S, T any](src *Node[S], upgrade func(*Node[T]) (*T), parent *Node[T]) *Node[T] {
	n := &Node[T]{
		Hof:      src.Hof,
		Value:     src.Value,
		Parent:   parent,
		Children: make([]*Node[T], 0, len(src.Children)),
	}

	n.T = upgrade(n)

	// walk, upgrading children
	for _, c := range src.Children {
		u := Upgrade(c, upgrade, n)
		n.Children = append(n.Children, u)
	}

	return n
}
