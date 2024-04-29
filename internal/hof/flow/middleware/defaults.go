package middleware

import (
	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"

	"github.com/opentofu/opentofu/internal/hof/flow/middleware/info"
	"github.com/opentofu/opentofu/internal/hof/flow/middleware/sync"
)

func UseDefaults(ctx *hofcontext.Context, opts flags.RootPflagpole, popts flags.FlowPflagpole) {
	// ctx.Use(dummy.NewDummy(opts, popts))
	// ctx.Use(info.NewPrint(opts, popts))
	ctx.Use(info.NewProgress(opts, popts))
	//ctx.Use(info.NewBookkeeping(info.BookkeepingConfig{
	//Workdir: ".hof/flow",
	//}, opts, popts))
	ctx.Use(sync.NewPool(opts, popts))
}
