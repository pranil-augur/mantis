package cueform

import (
	"fmt"

	"cuelang.org/go/cue"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type PrintArgsTask struct{}

func NewPrintArgsTask(val cue.Value) (hofcontext.Runner, error) {
	return &PrintArgsTask{}, nil
}

func (t *PrintArgsTask) Run(ctx *hofcontext.Context) (interface{}, error) {
	v := ctx.Value
	argsField := v.LookupPath(cue.ParsePath("args"))
	if argsField.Err() != nil {
		return nil, fmt.Errorf("error accessing args field: %v", argsField.Err())
	}
	fmt.Println(argsField)
	return nil, nil
}
