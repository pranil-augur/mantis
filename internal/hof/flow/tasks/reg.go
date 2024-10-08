/*
 * Augur AI Proprietary
 * Copyright (c) 2024 Augur AI, Inc.
 *
 * This file is licensed under the Augur AI Proprietary License.
 *
 * Attribution:
 * This work is based on code from https://github.com/hofstadter-io/hof, licensed under the Apache License 2.0.
 */

package tasks

import (
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"

	"github.com/opentofu/opentofu/internal/hof/flow/tasks/api"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/csp"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/cue"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/db"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/hof"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/kubernetes"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/kv"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/mantis"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/msg"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/opentf"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/os"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/prompt"
)

func RegisterDefaults(context *hofcontext.Context) {
	context.Register("noop", NewNoop)
	context.Register("nest", NewNest)

	context.Register("api.Call", api.NewCall)
	context.Register("api.Serve", api.NewServe)

	context.Register("csp.Chan", csp.NewChan)
	context.Register("csp.Recv", csp.NewRecv)
	context.Register("csp.Send", csp.NewSend)

	context.Register("cue.Format", cue.NewCueFormat)

	context.Register("db.Call", db.NewCall)
	context.Register("mantis.core.TF", opentf.NewTFTask)
	context.Register("mantis.core.K8s", kubernetes.NewK8sTask)
	context.Register("mantis.core.Evaluate", mantis.NewLocalEvaluator)
	context.Register("mantis.core.Relay", opentf.NewRelayTask)
	context.Register("hof.Template", hof.NewHofTemplate)

	context.Register("kv.Mem", kv.NewMem)

	context.Register("msg.IrcClient", msg.NewIrcClient)

	context.Register("os.Exec", os.NewExec)
	context.Register("os.FileLock", os.NewFileLock)
	context.Register("os.FileUnlock", os.NewFileUnlock)
	context.Register("os.Getenv", os.NewGetenv)
	context.Register("os.Glob", os.NewGlob)
	context.Register("os.Mkdir", os.NewMkdir)
	context.Register("os.ReadFile", os.NewReadFile)
	context.Register("os.ReadGlobs", os.NewReadGlobs)
	context.Register("os.Sleep", os.NewSleep)
	context.Register("os.Stdin", os.NewStdin)
	context.Register("os.Stdout", os.NewStdout)
	context.Register("os.Watch", os.NewWatch)
	context.Register("os.WriteFile", os.NewWriteFile)

	context.Register("prompt.Prompt", prompt.NewPrompt)

}
