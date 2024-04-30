/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package ast

type Node interface {
	Clone() Node
	CloneNodeBase() NodeBase

	Script() *Script

	DocLine() int
	SetDocLine(int)
	BegLine() int
	SetBegLine(int)
	EndLine() int
	SetEndLine(int)

	String() string
	Name() string
	Comment() string
	AddComment(string)
}
