package prompt

import (
	"cuelang.org/go/cue"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	libprompt "github.com/opentofu/opentofu/internal/hof/lib/prompt"
)

type Prompt struct{}

func NewPrompt(val cue.Value) (hofcontext.Runner, error) {
	return &Prompt{}, nil
}

// Tasks must implement a Run func, this is where we execute our task
func (T *Prompt) Run(ctx *hofcontext.Context) (any, error) {
	ctx.CUELock.Lock()
	defer ctx.CUELock.Unlock()

	r, err := libprompt.RunPrompt(ctx.Value)
	if err != nil {
		return nil, err
	}

	return r, nil
}
