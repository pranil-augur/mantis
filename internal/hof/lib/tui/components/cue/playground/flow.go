package playground

import (
	"bytes"

	"cuelang.org/go/cue"

	"github.com/opentofu/opentofu/internal/hof/cmd/hof/flags"
	flowcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/flow/middleware"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks"
	"github.com/opentofu/opentofu/internal/hof/flow/flow"
)


func (C *Playground) runFlow(val cue.Value) (cue.Value, error) {
	var stdin, stdout, stderr bytes.Buffer

	ctx := flowcontext.New()
	ctx.RootValue = val
	ctx.Stdin = &stdin
	ctx.Stdout = &stdout
	ctx.Stderr = &stderr

	// how to inject tags into original value
	// fill / return value
	middleware.UseDefaults(ctx, flags.RootPflagpole{}, flags.FlowPflagpole{})
	tasks.RegisterDefaults(ctx)

	f, err := flow.OldFlow(ctx, val)
	if err != nil {
		return val, err
	}

	C.flow = f

	err = f.Start()

	return f.Final, err
}
